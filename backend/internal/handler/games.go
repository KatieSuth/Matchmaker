package handler

import (
	"context"
	"errors"
	"log/slog"
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
		slog.WarnContext(c.Request.Context(), "request reached GetSystemGamesHandler without userID in context")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	systemGames, err := GetSystemGames(h.store, c.Request.Context())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to get system games", "user_id", userID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err,
		})
		return
	}

	slog.DebugContext(c.Request.Context(), "system games fetched", "user_id", userID, "count", len(systemGames))
	c.JSON(http.StatusOK, systemGames)
}

func GetSystemGames(s store.Store, ctx context.Context) ([]model.Game, error) {
	games, err := s.GetSystemGames(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "store error fetching system games", "error", err)
		return []model.Game{}, errors.New("Failed to fetch user")
	}
	return games, nil
}

// GET /games/:gameId/ranks
func (h *Handler) GetGameRanksByGame(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached GetGameRanksByGame without userID in context")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	gameId := c.Param("gameId")
	gameUUID, err := uuid.Parse(gameId)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid gameId param", "user_id", userID, "game_id", gameId, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "bad gameId",
		})
		return
	}

	gameRanks, err := GetGameRanksForGame(gameUUID, h.store, c.Request.Context())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to get game ranks", "user_id", userID, "game_id", gameUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err,
		})
		return
	}

	slog.DebugContext(c.Request.Context(), "game ranks fetched", "user_id", userID, "game_id", gameUUID, "count", len(gameRanks))
	c.JSON(http.StatusOK, gameRanks)
}

func GetGameRanksForGame(gameId uuid.UUID, s store.Store, ctx context.Context) ([]model.GameRank, error) {
	gameRanks, err := s.GetGameRanks(ctx, &gameId)
	if err != nil {
		slog.ErrorContext(ctx, "store error fetching game ranks", "game_id", gameId, "error", err)
		return []model.GameRank{}, errors.New("Failed to fetch user")
	}
	return gameRanks, nil
}
