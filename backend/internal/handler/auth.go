package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/apilink"
	"github.com/KatieSuth/MatchmakerAPI/internal/discord"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/textinput"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	if h.discord == nil {
		slog.ErrorContext(c.Request.Context(), "discord api not configured for OAuth callback")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not reach Discord API",
		})
		return
	}

	token, err := h.discord.Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "OAuth code exchange failed", "error", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	discordUser, err := h.discord.FetchMe(c.Request.Context(), token.AccessToken)
	if err != nil {
		if errors.Is(err, discord.ErrMalformedUser) {
			slog.ErrorContext(c.Request.Context(), "failed to decode Discord user response", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Unexpected, possibly malformed, response from Discord API",
			})
			return
		}
		slog.ErrorContext(c.Request.Context(), "failed to reach Discord API", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not reach Discord API",
		})
		return
	}

	user, err := h.upsertUserFromDiscordProfile(c.Request.Context(), discordUser)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to create new user", "discord_id", discordUser.ID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not locate the user account and could not create a new one",
		})
		return
	}

	if token.RefreshToken == "" {
		slog.ErrorContext(c.Request.Context(), "Discord OAuth response missing refresh token", "user_id", user.ID)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not store Discord authorization",
		})
		return
	}

	if err := h.apiLinkVault.PutRefreshToken(c.Request.Context(), user.ID, apilink.ProviderDiscord, token.RefreshToken); err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to persist Discord refresh token", "user_id", user.ID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not store Discord authorization",
		})
		return
	}
	slog.InfoContext(c.Request.Context(), "stored encrypted API refresh token", "user_id", user.ID, "provider", apilink.ProviderDiscord)

	if token.AccessToken != "" && !token.Expiry.IsZero() {
		h.discord.SeedAccessToken(user.ID, token.AccessToken, token.Expiry)
	}

	otc, err := h.issueOneTimeCode(c.Request.Context(), user.ID)
	if err != nil {
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

const testAuthBypassHeader = "X-Test-Auth-Bypass-Token"

// ShouldRegisterTestLogin is true only when the bypass flag is explicitly enabled and Gin is
// not in release mode. Used by main.go so the route cannot register in production even if the
// flag is accidentally set.
func ShouldRegisterTestLogin(ginMode, enabled string) bool {
	return enabled == "true" && ginMode != gin.ReleaseMode
}

// upsertUserFromDiscordProfile finds or creates the local user for a Discord profile and refreshes
// discord name/avatar on subsequent logins. Create failures are returned; update failures are
// logged and the existing user is still returned (matching the original OAuth callback behavior).
func (h *Handler) upsertUserFromDiscordProfile(ctx context.Context, discordUser model.DiscordUser) (model.User, error) {
	user, err := h.store.GetUserByDiscordID(ctx, discordUser.ID, true)
	if err != nil {
		slog.InfoContext(ctx, "user not found, creating new account", "discord_id", discordUser.ID)
		// Seed optional display_name from Discord global_name; never fail signup on normalize errors.
		var displayNamePtr *string
		rawGlobalName := ""
		if discordUser.GlobalName != nil {
			rawGlobalName = *discordUser.GlobalName
		}
		if normalized, normErr := textinput.NormalizeOptional(rawGlobalName, userDisplayNameMaxRunes); normErr == nil && normalized != "" {
			displayNamePtr = &normalized
		}
		user, err = h.store.CreateNewUser(ctx, discordUser, displayNamePtr)
		if err != nil {
			return model.User{}, err
		}
		slog.InfoContext(ctx, "new user created", "user_id", user.ID, "discord_id", discordUser.ID)
		return user, nil
	}

	if user.DiscordName != &discordUser.Username || user.ImageUrl != &discordUser.Avatar {
		slog.InfoContext(ctx, "updating user profile from Discord", "user_id", user.ID)
		updated, updateErr := h.store.UpdateUserFromLogin(ctx, user.ID, discordUser)
		if updateErr != nil {
			slog.ErrorContext(ctx, "failed to update user from Discord login", "user_id", user.ID, "error", updateErr)
			return user, nil
		}
		return updated, nil
	}
	return user, nil
}

// issueOneTimeCode persists a fresh OTC bound to userID and returns the plaintext code.
func (h *Handler) issueOneTimeCode(ctx context.Context, userID uuid.UUID) (string, error) {
	otcBytes := make([]byte, 16)
	_, _ = rand.Read(otcBytes)
	otc := hex.EncodeToString(otcBytes)
	if err := h.store.CreateOneTimeCode(ctx, otc, userID); err != nil {
		return "", err
	}
	return otc, nil
}

// POST /auth/test_login — test-only session mint used by Playwright. Never registered in
// GIN_MODE=release; still requires the shared-secret header even when the route exists.
func (h *Handler) TestLoginHandler(c *gin.Context) {
	if h.testAuthBypassToken == "" || h.ginMode == gin.ReleaseMode {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	got := c.GetHeader(testAuthBypassHeader)
	if got == "" || got != h.testAuthBypassToken {
		slog.WarnContext(c.Request.Context(), "test login rejected: missing or invalid bypass token")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var body struct {
		DiscordID  string  `json:"discord_id"`
		Username   string  `json:"username"`
		GlobalName *string `json:"global_name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "test login request bind failed", "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Improper json or json value types",
		})
		return
	}
	if strings.TrimSpace(body.DiscordID) == "" || strings.TrimSpace(body.Username) == "" {
		slog.WarnContext(c.Request.Context(), "test login request missing discord_id or username")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "discord_id and username are required",
		})
		return
	}

	discordUser := model.DiscordUser{
		ID:         body.DiscordID,
		Username:   body.Username,
		GlobalName: body.GlobalName,
	}
	user, err := h.upsertUserFromDiscordProfile(c.Request.Context(), discordUser)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "test login failed to upsert user", "discord_id", body.DiscordID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not locate the user account and could not create a new one",
		})
		return
	}

	otc, err := h.issueOneTimeCode(c.Request.Context(), user.ID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "test login failed to store one-time code", "user_id", user.ID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not store code to complete auth",
		})
		return
	}

	slog.InfoContext(c.Request.Context(), "test login issued one-time code", "user_id", user.ID, "new_user", user.NewUser)
	c.JSON(http.StatusOK, gin.H{
		"otc":      otc,
		"new_user": user.NewUser,
	})
}
