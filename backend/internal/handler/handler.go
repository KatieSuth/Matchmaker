// Package handler implements Matchmaker’s REST API on top of Gin: Discord OAuth, JWT auth,
// and JSON responses backed by the store.
package handler

import (
	"net/http"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"

	"golang.org/x/oauth2"
)

// Handler holds shared dependencies for all HTTP entrypoints. Mutating it after New is not supported.
type Handler struct {
	ginMode           string
	store             store.Store
	secureCookie      *securecookie.SecureCookie
	oauth2Config      *oauth2.Config
	cookieDomain      string
	frontendURL       string
	jwtSecret         []byte
	refreshExpiration int
	discordApiUrl     string
}

// New builds a Handler.
// gm is the Gin run mode; s is the persistence layer; sc signs/encrypts auth cookies;
// o2c is the Discord OAuth2 config; cd is the cookie domain; fURL is the frontend base URL;
// jwt is the JWT signing key; refExp is refresh-token Max-Age in seconds; dApi is Discord's API base URL.
func New(gm string, s store.Store, sc *securecookie.SecureCookie, o2c *oauth2.Config, cd string, fURL string, jwt []byte, refExp int, dApi string) *Handler {
	return &Handler{
		ginMode:           gm,
		store:             s,
		secureCookie:      sc,
		oauth2Config:      o2c,
		cookieDomain:      cd,
		frontendURL:       fURL,
		jwtSecret:         jwt,
		refreshExpiration: refExp,
		discordApiUrl:     dApi,
	}
}

// GET /health
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "API is running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
