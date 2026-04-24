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
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	systemGames, err := h.store.GetSystemGames(c.Request.Context())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to get system games", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch games",
		})
		return
	}

	c.JSON(http.StatusOK, systemGames)
}

// GET /games/users/:ownerId
func (h *Handler) GetUserGamesHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	ownerIDParam := c.Param("ownerId")
	ownerID, err := uuid.Parse(ownerIDParam)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid ownerId param", "user_id", userUUID, "owner_id", ownerIDParam, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "bad ownerId",
		})
		return
	}

	userGames, err := h.store.GetUserGames(c.Request.Context(), &ownerID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to get user games", "user_id", userUUID, "owner_id", ownerID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch user games",
		})
		return
	}

	c.JSON(http.StatusOK, userGames)
}

// GET /games/:gameId/ranks
func (h *Handler) GetGameRanksByGame(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	gameId := c.Param("gameId")
	gameUUID, err := uuid.Parse(gameId)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid gameId param", "user_id", userUUID, "game_id", gameId, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "bad gameId",
		})
		return
	}

	gameRanks, err := GetGameRanksForGame(gameUUID, h.store, c.Request.Context())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to get game ranks", "user_id", userUUID, "game_id", gameUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch game ranks",
		})
		return
	}

	c.JSON(http.StatusOK, gameRanks)
}

// GetGameRanksForGame loads rank tiers for a game; shared by the HTTP handler and tests.
func GetGameRanksForGame(gameId uuid.UUID, s store.Store, ctx context.Context) ([]model.GameRank, error) {
	gameRanks, err := s.GetGameRanks(ctx, &gameId)
	if err != nil {
		slog.ErrorContext(ctx, "store error fetching game ranks", "game_id", gameId, "error", err)
		return []model.GameRank{}, errors.New("Failed to fetch user")
	}
	return gameRanks, nil
}

// GET /games/:gameId/modes
func (h *Handler) GetGameModesByGame(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	gameId := c.Param("gameId")
	gameUUID, err := uuid.Parse(gameId)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid gameId param", "user_id", userUUID, "game_id", gameId, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "bad gameId",
		})
		return
	}

	gameModes, err := h.store.GetGameModes(c.Request.Context(), gameUUID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to get game modes", "user_id", userUUID, "game_id", gameUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch game modes",
		})
		return
	}

	c.JSON(http.StatusOK, gameModes)
}
