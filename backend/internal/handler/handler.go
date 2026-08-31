// Package handler implements Matchmaker’s REST API on top of Gin: Discord OAuth, JWT auth,
// and JSON responses backed by the store.
package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/apilink"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"

	"golang.org/x/oauth2"
)

// DiscordAPI is the Discord OAuth and REST subset handlers need for login and guild locks.
type DiscordAPI interface {
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	FetchMe(ctx context.Context, accessToken string) (model.DiscordUser, error)
	ListUserGuilds(ctx context.Context, userID uuid.UUID) ([]model.DiscordGuild, error)
	SeedAccessToken(userID uuid.UUID, accessToken string, expiry time.Time)
}

// Handler holds shared dependencies for all HTTP entrypoints. Mutating it after New is not supported.
type Handler struct {
	ginMode             string
	store               store.Store
	secureCookie        *securecookie.SecureCookie
	oauth2Config        *oauth2.Config
	generateState       func() (string, error)
	encodeStateCookie   func(name string, value interface{}) (string, error)
	cookieDomain        string
	frontendURL         string
	jwtSecret           []byte
	refreshExpiration   int
	discordApiUrl       string
	matchmakingSettings matchmaking.Settings
	apiLinkVault        *apilink.Vault
	discord             DiscordAPI
}

// New builds a Handler.
// gm is the Gin run mode; s is the persistence layer; sc signs/encrypts auth cookies;
// o2c is the Discord OAuth2 config; cd is the cookie domain; fURL is the frontend base URL;
// jwt is the JWT signing key; refExp is refresh-token Max-Age in seconds; dApi is Discord's API base URL;
// apiLinkKeys is the AES-256 keyring for third-party refresh tokens stored in api_links;
// discordAPI handles OAuth exchange, /users/@me, guild lists, and access-token cache seeding.
func New(gm string, s store.Store, sc *securecookie.SecureCookie, o2c *oauth2.Config, cd string, fURL string, jwt []byte, refExp int, dApi string, mmSettings matchmaking.Settings, apiLinkKeys *apilink.Keyring, discordAPI DiscordAPI) *Handler {
	return &Handler{
		ginMode:             gm,
		store:               s,
		secureCookie:        sc,
		oauth2Config:        o2c,
		generateState:       GenerateState,
		encodeStateCookie:   sc.Encode,
		cookieDomain:        cd,
		frontendURL:         fURL,
		jwtSecret:           jwt,
		refreshExpiration:   refExp,
		discordApiUrl:       dApi,
		matchmakingSettings: mmSettings,
		apiLinkVault:        apilink.New(apiLinkKeys, s),
		discord:             discordAPI,
	}
}

// userIDFromContext returns the user UUID set by JWT middleware, or writes an error response and false.
func userIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached handler without userID in context")
		c.AbortWithStatus(http.StatusUnauthorized)
		return uuid.Nil, false
	}

	var userUUID uuid.UUID
	switch typedUserID := userID.(type) {
	case string:
		parsed, err := uuid.Parse(typedUserID)
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "failed to parse userID into UUID", "user_id", userID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Could not parse user ID",
			})
			return uuid.Nil, false
		}
		userUUID = parsed
	case uuid.UUID:
		userUUID = typedUserID
	default:
		slog.ErrorContext(c.Request.Context(), "failed to parse userID into UUID", "user_id", userID, "error", "unexpected userID type")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not parse user ID",
		})
		return uuid.Nil, false
	}
	return userUUID, true
}

// GET /health
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "API is running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
