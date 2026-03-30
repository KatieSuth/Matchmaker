package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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
	log.Print("top of discord callback")
	cookieVal, err := c.Cookie("oauth_state")

	if err != nil {
		log.Printf("error getting cookie: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//decode & verify the signature
	log.Print("getting state signature")
	var storedState string
	if err := h.secureCookie.Decode("oauth_state", cookieVal, &storedState); err != nil {
		log.Printf("error decoding stored state: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized) //oauth_state is either tampered with or expired
		return
	}

	//delete the cookie immediately, it's one-time use
	log.Print("deleting cookie")
	c.SetCookie("oauth_state", "", -1, "/", "", true, true)

	log.Print("checking query state matches stored state")
	if c.Query("state") != storedState {
		log.Printf("query state did not match stored state: %v, %v", c.Query("state"), storedState)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//exchange the code for the user's token
	log.Print("getting user token from code")
	token, err := h.oauth2Config.Exchange(c, c.Query("code"))
	if err != nil {
		log.Printf("could not exchange code for token: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//use the token to fetch the user from Discord
	log.Print("getting user from discord API")
	client := h.oauth2Config.Client(c.Request.Context(), token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		log.Printf("could not get user from API: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var user model.DiscordUser
	log.Print("decoding Discord response")
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Printf("could not deocde discord response: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, user)
}
