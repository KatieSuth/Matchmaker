package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/textinput"
	"github.com/gin-gonic/gin"
)

// OAuthStateCookieMaxAge is the Discord OAuth CSRF cookie lifetime in seconds.
// Sized to tolerate Cloud Run cold starts during the Discord round-trip.
const OAuthStateCookieMaxAge = 900

// generateAccessTokenForRefresh is swappable in tests to cover refresh error paths.
var generateAccessTokenForRefresh = model.GenerateAccessToken

// GenerateState returns a random hex string used as the OAuth2 state parameter to bind the
// callback to the user’s original /auth/login request (CSRF protection).
func GenerateState() (string, error) {
	state := make([]byte, 16)
	_, err := rand.Read(state)
	return hex.EncodeToString(state), err
}

// hashToken returns a stable SHA-256 digest of a bearer or refresh value; we persist only
// the hash in the database so a DB leak does not immediately expose raw tokens.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// sessionCookieDomains lists Domain attribute values to clear. Browsers can retain legacy
// host-only and domain-scoped cookies side by side; clearing every variant avoids stale
// refresh_token values shadowing the current session cookie.
func (h *Handler) sessionCookieDomains() []string {
	if h.cookieDomain == "" {
		return []string{""}
	}
	return []string{h.cookieDomain, ""}
}

// writeSessionCookie sets a secure session cookie, optionally scoped to the configured domain.
func (h *Handler) writeSessionCookie(c *gin.Context, name, value string, maxAge int, httpOnly bool, domain string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
	}
	if domain != "" {
		cookie.Domain = domain
	}
	http.SetCookie(c.Writer, cookie)
}

// clearSessionCookies expires refresh_token and auth_session for every domain variant the app may have used.
func (h *Handler) clearSessionCookies(c *gin.Context) {
	for _, domain := range h.sessionCookieDomains() {
		h.writeSessionCookie(c, "refresh_token", "", -1, true, domain)
		h.writeSessionCookie(c, "auth_session", "", -1, false, domain)
	}
}

// setAuthCookies sets the HttpOnly refresh_token and a lightweight auth_session flag the
// Next.js layer uses for client-side route guards (API calls still require a valid JWT).
func (h *Handler) setAuthCookies(c *gin.Context, refreshToken string, maxAge int) {
	h.clearSessionCookies(c)
	if maxAge < 0 || refreshToken == "" {
		return
	}

	h.writeSessionCookie(c, "refresh_token", refreshToken, maxAge, true, h.cookieDomain)
	h.writeSessionCookie(c, "auth_session", "1", maxAge, false, h.cookieDomain)
}

// GET /auth/login
func (h *Handler) LoginHandler(c *gin.Context) {
	//generate state for the Discord
	state, err := h.generateState()
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to generate OAuth state", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	encoded, err := h.encodeStateCookie("oauth_state", state)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to generate OAuth state cookie", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.SetCookie("oauth_state", encoded, OAuthStateCookieMaxAge, "/", h.cookieDomain, true, true)
	c.Redirect(http.StatusTemporaryRedirect, h.oauth2Config.AuthCodeURL(state))
}

// GET /auth/discord_redirect
func (h *Handler) DiscordCallbackHandler(c *gin.Context) {
	cookieVal, err := c.Cookie("oauth_state")

	if err != nil {
		slog.WarnContext(c.Request.Context(), "OAuth callback missing state cookie")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//decode & verify the signature
	var storedState string
	if err := h.secureCookie.Decode("oauth_state", cookieVal, &storedState); err != nil {
		slog.WarnContext(c.Request.Context(), "OAuth state cookie failed decode — possibly tampered or expired", "error", err)
		c.AbortWithStatus(http.StatusUnauthorized) //oauth_state is either tampered with or expired
		return
	}

	//delete the cookie immediately, it's one-time use
	c.SetCookie("oauth_state", "", -1, "/", h.cookieDomain, true, true)

	if c.Query("state") != storedState {
		slog.WarnContext(c.Request.Context(), "OAuth state mismatch — possible CSRF attempt")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//exchange the code for the user's token
	token, err := h.oauth2Config.Exchange(c, c.Query("code"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "OAuth code exchange failed", "error", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//use the token to fetch the user from Discord
	client := h.oauth2Config.Client(c.Request.Context(), token)
	resp, err := client.Get(h.discordApiUrl + "/users/@me")
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to reach Discord API", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not reach Discord API",
		})
		return
	}
	defer resp.Body.Close()

	var discordUser model.DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to decode Discord user response", "error", err)
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
		slog.InfoContext(c.Request.Context(), "user not found, creating new account", "discord_id", discordUser.ID)
		// Seed optional display_name from Discord global_name; never fail signup on normalize errors.
		var displayNamePtr *string
		rawGlobalName := ""
		if discordUser.GlobalName != nil {
			rawGlobalName = *discordUser.GlobalName
		}
		if normalized, normErr := textinput.NormalizeOptional(rawGlobalName, userDisplayNameMaxRunes); normErr == nil && normalized != "" {
			displayNamePtr = &normalized
		}
		user, err = h.store.CreateNewUser(c.Request.Context(), discordUser, displayNamePtr)
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "failed to create new user", "discord_id", discordUser.ID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Could not locate the user account and could not create a new one",
			})
			return
		}
		slog.InfoContext(c.Request.Context(), "new user created", "user_id", user.ID, "discord_id", discordUser.ID)
	} else {
		//found the user, update them if necessary
		if user.DiscordName != &discordUser.Username || user.ImageUrl != &discordUser.Avatar {
			slog.InfoContext(c.Request.Context(), "updating user profile from Discord", "user_id", user.ID)
			user, err = h.store.UpdateUserFromLogin(c.Request.Context(), user.ID, discordUser)
			if err != nil {
				slog.ErrorContext(c.Request.Context(), "failed to update user from Discord login", "user_id", user.ID, "error", err)
			}
		}
	}

	//generate a short-lived one-time code to exchange for tokens in /auth/complete
	otcBytes := make([]byte, 16)
	rand.Read(otcBytes)
	otc := hex.EncodeToString(otcBytes)

	if err := h.store.CreateOneTimeCode(c.Request.Context(), otc, user.ID); err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to store one-time code", "user_id", user.ID, "error", err)
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
		slog.WarnContext(c.Request.Context(), "missing refresh token cookie")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//parse and validate refresh token
	refreshHashStr := hashToken(cookieVal)

	refresh, err := h.store.GetRefreshToken(c.Request.Context(), refreshHashStr)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "refresh token not found in database", "error", err)
		h.clearSessionCookies(c)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if time.Now().After(refresh.ExpiresAt) {
		slog.InfoContext(c.Request.Context(), "refresh token expired, deleting", "user_id", refresh.UserID, "expired_at", refresh.ExpiresAt)
		//refresh token has expired, delete it
		err = h.store.DeleteRefreshToken(c.Request.Context(), refreshHashStr)
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "failed to delete expired refresh token", "user_id", refresh.UserID, "error", err)
		}

		//clear the cookies
		h.setAuthCookies(c, "", -1)

		//return unauthorized
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Issue a new access token only. Keep the existing refresh token so concurrent refresh
	// requests (e.g. React Strict Mode) cannot invalidate a still-valid browser cookie.
	accessToken, err := generateAccessTokenForRefresh(refresh.UserID.String(), h.jwtSecret)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to generate access token during refresh", "user_id", refresh.UserID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not refresh required tokens",
		})
		return
	}

	// Re-set session cookies to extend Max-Age without rotating the refresh token value.
	h.setAuthCookies(c, cookieVal, h.refreshExpiration)

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

	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "auth complete request bind failed", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Improper json or json value types",
		})
		return
	}

	if body.OTC == "" {
		slog.WarnContext(c.Request.Context(), "auth complete request missing OTC")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "otc is required",
		})
		return
	}

	//look up and immediately delete the one time code
	otcUserID, err := h.store.ConsumeOneTimeCode(c.Request.Context(), body.OTC)
	if err != nil || otcUserID.String() == "" {
		slog.WarnContext(c.Request.Context(), "failed to consume one-time code — invalid or already used", "error", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//generate tokens
	accessToken, refreshToken, err := model.GenerateTokens(otcUserID.String(), h.jwtSecret)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to generate tokens after OTC exchange", "user_id", otcUserID, "error", err)
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
		slog.ErrorContext(c.Request.Context(), "failed to store refresh token after OTC exchange", "user_id", otcUserID, "error", err)
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
	cookieVal, err := c.Cookie("refresh_token")
	if err != nil {
		//No cookie, already logged out
		//calling this anyway to make sure auth_session is deleted too
		h.setAuthCookies(c, "", -1)
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	// hash the token, look it up in DB & delete
	tokenHash := hashToken(cookieVal)

	if err := h.store.DeleteRefreshToken(c.Request.Context(), tokenHash); err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to delete refresh token on logout", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	h.setAuthCookies(c, "", -1)
	c.AbortWithStatus(http.StatusNoContent)
}
