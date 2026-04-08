package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/gin-gonic/gin"
)

func generateState() (string, error) {
	state := make([]byte, 16)
	_, err := rand.Read(state)
	return hex.EncodeToString(state), err
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (h *Handler) setAuthCookies(c *gin.Context, refreshToken string, maxAge int) {
	c.SetCookie("refresh_token", refreshToken, maxAge, "/", h.cookieDomain, true, true)
	c.SetCookie("auth_session", "1", maxAge, "/", h.cookieDomain, true, false)
}

// GET /auth/login
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

	c.SetCookie("oauth_state", encoded, 300, "/", h.cookieDomain, true, true)
	c.Redirect(http.StatusTemporaryRedirect, h.oauth2Config.AuthCodeURL(state))
}

// GET /auth/discord_callback
func (h *Handler) DiscordCallbackHandler(c *gin.Context) {
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
	c.SetCookie("oauth_state", "", -1, "/", h.cookieDomain, true, true)

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

	//handle local storage for the user
	var user model.User
	user, err = h.store.GetUserByDiscordID(c.Request.Context(), discordUser.ID, true)
	if err != nil {
		//couldn't find the user, create them
		user, err = h.store.CreateNewUser(c.Request.Context(), discordUser)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Could not locate the user account and could not create a new one",
			})
			return
		}
	} else {
		//found the user, update them if necessary
		if user.DiscordName != &discordUser.Username || user.ImageUrl != &discordUser.Avatar {
			user, err = h.store.UpdateUserFromLogin(c.Request.Context(), user.ID, discordUser)
		}
	}

	//generate a short-lived one-time code to exchange for tokens in /auth/complete
	otcBytes := make([]byte, 16)
	rand.Read(otcBytes)
	otc := hex.EncodeToString(otcBytes)

	if err := h.store.CreateOneTimeCode(c.Request.Context(), otc, user.ID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not store code to complete auth",
		})
		return
	}

	redirectURL := fmt.Sprintf("%s/auth/callback?&otc=%s&new_user=%t", h.frontendURL, otc, user.NewUser)
	c.Redirect(http.StatusFound, redirectURL)
}

// POST /auth/refresh
func (h *Handler) RefreshHandler(c *gin.Context) {
	cookieVal, err := c.Cookie("refresh_token")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//parse and validate refresh token
	refreshHashStr := hashToken(cookieVal)

	refresh, err := h.store.GetRefreshToken(c.Request.Context(), refreshHashStr)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if time.Now().After(refresh.ExpiresAt) {
		//refresh token has expired, delete it
		err = h.store.DeleteRefreshToken(c.Request.Context(), refreshHashStr)
		if err != nil {
			log.Printf("Could not delete expired refresh token; %v", err)
		}

		//clear the cookies
		h.setAuthCookies(c, "", -1)

		//return unauthorized
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//generate fresh tokens
	accessToken, refreshToken, err := model.GenerateTokens(refresh.UserID.String(), h.jwtSecret)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not refresh required tokens",
		})
		return
	}

	//store the new refresh token
	refreshHashNewStr := hashToken(refreshToken)

	expireTime := time.Now().Add(time.Duration(h.refreshExpiration) * time.Second)
	_, err = h.store.CreateNewRefreshToken(c.Request.Context(), refreshHashNewStr, refresh.UserID, expireTime)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not store required tokens",
		})
		return
	}

	//set refresh token that lasts the specific amount of time (default: 7 days) in HttpOnly cookie
	h.setAuthCookies(c, refreshToken, h.refreshExpiration)

	//delete the old refresh token
	err = h.store.DeleteRefreshToken(c.Request.Context(), refreshHashStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not rotate the refresh token",
		})
		return
	}

	//return token
	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

// POST /auth/complete
func (h *Handler) CompleteAuthHandler(c *gin.Context) {
	var body struct {
		OTC string `json:"otc"`
	}

	err := c.ShouldBindJSON(&body)

	if err != nil || body.OTC == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "otc is required",
		})
		return
	}

	//look up and immediately delete the one time code
	otcUserID, err := h.store.ConsumeOneTimeCode(c.Request.Context(), body.OTC)
	if err != nil || otcUserID.String() == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//generate tokens
	accessToken, refreshToken, err := model.GenerateTokens(otcUserID.String(), h.jwtSecret)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not generate required tokens",
		})
		return
	}

	//hash & store the refresh token
	refreshHashStr := hashToken(refreshToken)

	expireTime := time.Now().Add(time.Duration(h.refreshExpiration) * time.Second)
	_, err = h.store.CreateNewRefreshToken(c.Request.Context(), refreshHashStr, otcUserID, expireTime)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not store required tokens",
		})
		return
	}

	h.setAuthCookies(c, refreshToken, h.refreshExpiration)
	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

// POST /auth/logout
func (h *Handler) LogoutHandler(c *gin.Context) {
	//TODO: delete refresh_token db value & cookie, delete auth_session cookie
	cookieVal, err := c.Cookie("refresh_token")
	if err != nil {
		//No cookie, already logged out
		c.Status(http.StatusNoContent)
		return
	}

	// hash the token, look it up in DB & delete
	tokenHash := hashToken(cookieVal)

	if err := h.store.DeleteRefreshToken(c.Request.Context(), tokenHash); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	h.setAuthCookies(c, "", -1)
	c.Status(http.StatusNoContent)
}
