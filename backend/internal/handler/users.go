package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
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

	user, err := GetUserById(userID.(string), h.store, c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err,
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetUserById(userId string, s store.Store, ctx context.Context) (model.User, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		log.Printf("Error parsing user ID string (%s) into UUID: %v", userId, err)
		return model.User{}, errors.New("Could now parse user ID")
	}

	user, err := s.GetUserByUserID(ctx, userUUID)
	if err != nil {
		log.Printf("Error fetching user (%s): %v", userUUID.String(), err)
		return model.User{}, errors.New("Failed to fetch user")
	}
	return user, nil
}

// GET /users/me/games
func (h *Handler) UsersMeGamesHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("UserID: %v, exists, %v", userID, exists)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userGames, err := GetUserGames(userID.(string), h.store, c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err,
		})
		return
	}

	c.JSON(http.StatusOK, userGames)
}

func GetUserGames(userId string, s store.Store, ctx context.Context) ([]model.UserGame, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		log.Printf("Error parsing user ID string (%s) into UUID: %v", userId, err)
		return []model.UserGame{}, errors.New("Could now parse user ID")
	}

	userGames, err := s.GetUserGamesForUser(ctx, userUUID)
	if err != nil {
		log.Printf("Error fetching user (%s): %v", userUUID.String(), err)
		return []model.UserGame{}, errors.New("Failed to fetch user")
	}
	return userGames, nil
}
