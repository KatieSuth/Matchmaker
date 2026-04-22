package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/store"
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

type updateEventGroupSettingsRequest struct {
	Region string `json:"region"`
	SubMin int32  `json:"sub_min"`
}

type updateRegistrationStatusRequest struct {
	RegistrationOpen bool `json:"registration_open"`
}

type upsertRegistrationRequest struct {
	CanSubstitute bool   `json:"can_substitute"`
	CanLobbyHost  bool   `json:"can_lobby_host"`
	DuoRequest    string `json:"duo_request"`
}

func userIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached handler without userID in context")
		c.AbortWithStatus(http.StatusUnauthorized)
		return uuid.Nil, false
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to parse userID into UUID", "user_id", userID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not parse user ID",
		})
		return uuid.Nil, false
	}
	return userUUID, true
}

// POST /events
func (h *Handler) CreateEventHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
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

// GET /events/:groupId
func (h *Handler) GetEventGroupHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid groupId in GetEventGroupHandler", "user_id", userUUID, "group_id", c.Param("groupId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "groupId must be a valid UUID",
		})
		return
	}

	detail, err := h.store.GetEventGroupDetail(c.Request.Context(), groupID, userUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(c.Request.Context(), "event group not found", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "failed to fetch event group detail", "group_id", groupID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch event group"})
		return
	}

	slog.InfoContext(c.Request.Context(), "event group fetched", "user_id", userUUID, "group_id", groupID, "events_count", len(detail.Events))
	c.JSON(http.StatusOK, detail)
}

// PATCH /events/:groupId
func (h *Handler) UpdateEventGroupSettingsHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid groupId in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", c.Param("groupId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "groupId must be a valid UUID"})
		return
	}

	var body updateEventGroupSettingsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid request body in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", groupID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Improper json or json value types"})
		return
	}

	err = h.store.UpdateEventGroupSettings(c.Request.Context(), groupID, userUUID, body.Region, body.SubMin)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden event group settings update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can edit this event group"})
		case errors.Is(err, store.ErrInvalidSubMin):
			slog.WarnContext(c.Request.Context(), "invalid event group settings payload", "user_id", userUUID, "group_id", groupID, "region", body.Region, "sub_min", body.SubMin)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "region is required and sub_min must be >= 0"})
		case errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for settings update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to update event group settings", "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update event group"})
		}
		return
	}

	slog.InfoContext(c.Request.Context(), "event group settings updated", "user_id", userUUID, "group_id", groupID, "region", body.Region, "sub_min", body.SubMin)
	c.Status(http.StatusNoContent)
}

// DELETE /events/:groupId
func (h *Handler) DeleteEventGroupHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid groupId in DeleteEventGroupHandler", "user_id", userUUID, "group_id", c.Param("groupId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "groupId must be a valid UUID"})
		return
	}

	err = h.store.DeleteEventGroup(c.Request.Context(), groupID, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden event group delete attempt", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can delete this event group"})
		case errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for delete", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to delete event group", "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to delete event group"})
		}
		return
	}

	slog.InfoContext(c.Request.Context(), "event group deleted", "user_id", userUUID, "group_id", groupID)
	c.Status(http.StatusNoContent)
}

// PATCH /events/:groupId/registration
func (h *Handler) UpdateEventGroupRegistrationStatusHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid groupId in UpdateEventGroupRegistrationStatusHandler", "user_id", userUUID, "group_id", c.Param("groupId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "groupId must be a valid UUID"})
		return
	}

	var body updateRegistrationStatusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid request body in UpdateEventGroupRegistrationStatusHandler", "user_id", userUUID, "group_id", groupID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Improper json or json value types"})
		return
	}

	err = h.store.SetEventGroupRegistrationOpen(c.Request.Context(), groupID, userUUID, body.RegistrationOpen)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden registration status update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can change registration status"})
		case errors.Is(err, store.ErrOpenRegistrationTeams):
			slog.WarnContext(c.Request.Context(), "registration open blocked due to existing teams", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Delete teams before opening registration"})
		case errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for registration status update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to update registration status", "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update registration status"})
		}
		return
	}

	slog.InfoContext(c.Request.Context(), "event group registration status updated", "user_id", userUUID, "group_id", groupID, "registration_open", body.RegistrationOpen)
	c.Status(http.StatusNoContent)
}

// POST /events/:groupId/teams
func (h *Handler) CreateTeamsHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid groupId in CreateTeamsHandler", "user_id", userUUID, "group_id", c.Param("groupId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "groupId must be a valid UUID"})
		return
	}

	err = h.store.CreateTeamsForGroup(c.Request.Context(), groupID, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden create teams attempt", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can create teams"})
		case errors.Is(err, store.ErrTeamsAlreadyCreated):
			slog.WarnContext(c.Request.Context(), "create teams blocked because teams already exist", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Teams already exist for this event group"})
		case errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for create teams", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to create teams", "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create teams"})
		}
		return
	}
	slog.InfoContext(c.Request.Context(), "teams created for event group", "user_id", userUUID, "group_id", groupID)
	c.Status(http.StatusNoContent)
}

// DELETE /events/:groupId/teams
func (h *Handler) DeleteTeamsAndOpenRegistrationHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid groupId in DeleteTeamsAndOpenRegistrationHandler", "user_id", userUUID, "group_id", c.Param("groupId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "groupId must be a valid UUID"})
		return
	}

	err = h.store.DeleteTeamsAndOpenRegistration(c.Request.Context(), groupID, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden delete teams attempt", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can delete teams"})
		case errors.Is(err, store.ErrTeamsNotCreated):
			slog.WarnContext(c.Request.Context(), "delete teams requested but no teams exist", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "No teams exist for this event group"})
		case errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for delete teams", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to delete teams and open registration", "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to delete teams"})
		}
		return
	}
	slog.InfoContext(c.Request.Context(), "teams deleted and registration reopened", "user_id", userUUID, "group_id", groupID)
	c.Status(http.StatusNoContent)
}

// PUT /registrations/:eventId/me
func (h *Handler) UpsertMyRegistrationHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}
	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid eventId in UpsertMyRegistrationHandler", "user_id", userUUID, "event_id", c.Param("eventId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "eventId must be a valid UUID"})
		return
	}

	var body upsertRegistrationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid request body in UpsertMyRegistrationHandler", "user_id", userUUID, "event_id", eventID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Improper json or json value types"})
		return
	}

	var duoRequest *string
	if trimmed := strings.TrimSpace(body.DuoRequest); trimmed != "" {
		duoRequest = &trimmed
	}

	err = h.store.UpsertRegistrationForEvent(c.Request.Context(), eventID, userUUID, body.CanSubstitute, body.CanLobbyHost, duoRequest)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrRegistrationClosed):
			slog.WarnContext(c.Request.Context(), "registration upsert rejected because registration is closed", "user_id", userUUID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Registration is closed"})
		case errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event not found for registration upsert", "user_id", userUUID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to upsert registration", "event_id", eventID, "user_id", userUUID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to save registration"})
		}
		return
	}

	slog.InfoContext(c.Request.Context(), "registration upserted", "user_id", userUUID, "event_id", eventID, "can_substitute", body.CanSubstitute, "can_lobby_host", body.CanLobbyHost)
	c.Status(http.StatusNoContent)
}

// DELETE /registrations/:eventId/me
// DELETE /registrations/:eventId/:userId
func (h *Handler) DeleteRegistrationHandler(c *gin.Context) {
	actorID, ok := userIDFromContext(c)
	if !ok {
		return
	}
	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid eventId in DeleteRegistrationHandler", "user_id", actorID, "event_id", c.Param("eventId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "eventId must be a valid UUID"})
		return
	}
	targetUserID := actorID
	if rawTarget := strings.TrimSpace(c.Param("userId")); rawTarget != "" {
		parsedTarget, err := uuid.Parse(rawTarget)
		if err != nil {
			slog.WarnContext(c.Request.Context(), "invalid userId in DeleteRegistrationHandler", "actor_user_id", actorID, "event_id", eventID, "user_id", rawTarget, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "userId must be a valid UUID"})
			return
		}
		targetUserID = parsedTarget
	}

	err = h.store.DeleteRegistrationForEvent(c.Request.Context(), eventID, targetUserID, actorID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden registration delete attempt", "actor_user_id", actorID, "event_id", eventID, "target_user_id", targetUserID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "You do not have permission to delete this registration"})
		case errors.Is(err, store.ErrRegistrationClosed):
			slog.WarnContext(c.Request.Context(), "registration delete rejected because registration is closed", "actor_user_id", actorID, "event_id", eventID, "target_user_id", targetUserID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Registration is closed"})
		case errors.Is(err, store.ErrRegistrationNotFound):
			slog.WarnContext(c.Request.Context(), "registration not found for delete", "actor_user_id", actorID, "event_id", eventID, "target_user_id", targetUserID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Registration not found"})
		case errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event not found for registration delete", "actor_user_id", actorID, "event_id", eventID, "target_user_id", targetUserID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to delete registration", "event_id", eventID, "target_user_id", targetUserID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to delete registration"})
		}
		return
	}

	slog.InfoContext(c.Request.Context(), "registration deleted", "actor_user_id", actorID, "event_id", eventID, "target_user_id", targetUserID)
	c.Status(http.StatusNoContent)
}
