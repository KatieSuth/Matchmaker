package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// POST /users/me
func (h *Handler) UsersMeHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("UserID: %v, exists, %v", userID, exists)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not parse user ID",
		})
		return
	}

	user, err := h.store.GetUserByUserID(c.Request.Context(), userUUID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch user",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}
