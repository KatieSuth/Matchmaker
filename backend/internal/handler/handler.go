package handler

import (
	"net/http"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"

	"golang.org/x/oauth2"
)

type Handler struct {
	ginMode           string
	store             store.Store
	secureCookie      *securecookie.SecureCookie
	oauth2Config      *oauth2.Config
	cookieDomain      string
	frontendURL       string
	jwtSecret         []byte
	refreshExpiration int
}

func New(gm string, s store.Store, sc *securecookie.SecureCookie, o2c *oauth2.Config, cd string, fURL string, jwt []byte, refExp int) *Handler {
	return &Handler{
		ginMode:           gm,
		store:             s,
		secureCookie:      sc,
		oauth2Config:      o2c,
		cookieDomain:      cd,
		frontendURL:       fURL,
		jwtSecret:         jwt,
		refreshExpiration: refExp,
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
