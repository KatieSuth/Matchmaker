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

// GET /games
func (h *Handler) GetSystemGamesHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("UserID: %v, exists, %v", userID, exists)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	systemGames, err := GetSystemGames(h.store, c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err,
		})
		return
	}

	c.JSON(http.StatusOK, systemGames)
}

func GetSystemGames(s store.Store, ctx context.Context) ([]model.Game, error) {
	games, err := s.GetSystemGames(ctx)
	if err != nil {
		log.Printf("Error fetching system games: %v", err)
		return []model.Game{}, errors.New("Failed to fetch user")
	}
	return games, nil
}

// GET /games/:gameId/ranks
func (h *Handler) GetGameRanksByGame(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("UserID: %v, exists, %v", userID, exists)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	gameId := c.Param("gameId")
	gameUUID, err := uuid.Parse(gameId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "bad gameId",
		})
	}

	gameRanks, err := GetGameRanksForGame(gameUUID, h.store, c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err,
		})
		return
	}

	c.JSON(http.StatusOK, gameRanks)
}

func GetGameRanksForGame(gameId uuid.UUID, s store.Store, ctx context.Context) ([]model.GameRank, error) {
	gameRanks, err := s.GetGameRanks(ctx, &gameId)
	if err != nil {
		log.Printf("Error fetching system games: %v", err)
		return []model.GameRank{}, errors.New("Failed to fetch user")
	}
	return gameRanks, nil
}
