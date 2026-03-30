package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/gin-gonic/gin"
)

func generateState() (string, error) {
	state := make([]byte, 16)
	_, err := rand.Read(state)
	return hex.EncodeToString(state), err
}

// POST /login
func (h *Handler) LoginHandler(c *gin.Context) {
	//generate state for the Discord
	state, err := generateState()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	encoded, err := h.secureCookie.Encode("oauth_state", state)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.SetCookie("oauth_state", encoded, 300, "/", "", true, true)
	c.Redirect(http.StatusTemporaryRedirect, h.oauth2Config.AuthCodeURL(state))
}

// POST /discord_callback
func (h *Handler) DiscordCallback(c *gin.Context) {
	cookieVal, err := c.Cookie("oauth_state")

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//decode & verify the signature
	var storedState string
	if err := h.secureCookie.Decode("oauth_state", cookieVal, &storedState); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized) //oauth_state is either tampered with or expired
		return
	}

	//delete the cookie immediately, it's one-time use
	c.SetCookie("oauth_state", "", -1, "/", "", true, true)

	if c.Query("state") != storedState {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//exchange the code for the user's token
	token, err := h.oauth2Config.Exchange(c, c.Query("code"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//use the token to fetch the user from Discord
	client := h.oauth2Config.Client(c.Request.Context(), token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not reach Discord API",
		})
		return
	}
	defer resp.Body.Close()

	var discordUser model.DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Unexpected, possibly malformed, response from Discord API",
		})
		return
	}

	var user model.User
	user, err = h.store.GetUserByDiscordID(c.Request.Context(), discordUser.ID, true)
	log.Printf("Check for user error: %v", err)
	if err != nil {
		//couldn't find the user, create them
		user, err = h.store.CreateNewUser(c.Request.Context(), discordUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Could not locate the user account and could not create a new one, check logs",
			})
			return
		}
	} else {
		//found the user, update them if necessary
		if user.DiscordName != &discordUser.Username || user.ImageUrl != &discordUser.Avatar {
			user, err = h.store.UpdateUserFromLogin(c.Request.Context(), user.ID, discordUser)
		}
	}

	//TODO: generate authentication tokens

	if user.NewUser {
		//redirect to my_account
		target := fmt.Sprintf("%s/my_account", h.frontendURL)
		c.Redirect(http.StatusTemporaryRedirect, target)
	} else {
		//redirect to events
		target := fmt.Sprintf("%s/events", h.frontendURL)
		c.Redirect(http.StatusTemporaryRedirect, target)
	}
}
