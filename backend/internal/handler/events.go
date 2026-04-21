package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type createEventRequest struct {
	GameModeID       string `json:"game_mode_id"`
	Region           string `json:"region"`
	StartTime        string `json:"start_time"`
	SubMin           int32  `json:"sub_min"`
	GamesToRun       int32  `json:"games_to_run"`
	RegistrationOpen bool   `json:"registration_open"`
}

// POST /events
func (h *Handler) CreateEventHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached CreateEventHandler without userID in context")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to parse userID into UUID", "user_id", userID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not parse user ID",
		})
		return
	}

	var body createEventRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Improper json or json value types",
		})
		return
	}

	gameModeID, err := uuid.Parse(body.GameModeID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "game_mode_id must be a valid UUID",
		})
		return
	}

	region := strings.TrimSpace(body.Region)
	if region == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "region is required",
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_time must be RFC3339 datetime",
		})
		return
	}

	if startTime.Before(time.Now()) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_time cannot be in the past",
		})
		return
	}

	if body.SubMin < 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "sub_min cannot be negative",
		})
		return
	}

	if body.GamesToRun <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "games_to_run must be greater than 0",
		})
		return
	}

	groupID, err := h.store.CreateEventGroupWithEvents(
		c.Request.Context(),
		userUUID,
		gameModeID,
		body.SubMin,
		body.RegistrationOpen,
		region,
		startTime,
		body.GamesToRun,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "game mode not found",
			})
			return
		}

		slog.ErrorContext(c.Request.Context(), "failed to create event group", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Something went wrong",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"group_id": groupID.String(),
	})
}
