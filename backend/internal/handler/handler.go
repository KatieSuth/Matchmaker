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

/*
// GET /api/v1/items
func (h *Handler) ListItems(c *gin.Context) {
	items := h.store.List()
	if items == nil {
		items = []model.Item{}
	}
	c.JSON(http.StatusOK, items)
}

// POST /api/v1/items
func (h *Handler) CreateItem(c *gin.Context) {
	var req model.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item := h.store.Create(req)
	c.JSON(http.StatusCreated, item)
}

// GET /api/v1/items/:id
func (h *Handler) GetItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	item, err := h.store.Get(id)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// PUT /api/v1/items/:id
func (h *Handler) UpdateItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req model.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.store.Update(id, req)
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// DELETE /api/v1/items/:id
func (h *Handler) DeleteItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.store.Delete(id); errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
*/
