package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/textinput"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const eventGroupNameMaxRunes = 50

type createEventRequest struct {
	GameModeID       string `json:"game_mode_id"`
	Region           string `json:"region"`
	StartTime        string `json:"start_time"`
	SubMin           int32  `json:"sub_min"`
	GamesToRun       int32  `json:"games_to_run"`
	RegistrationOpen bool   `json:"registration_open"`
	SortLogic        string `json:"sort_logic"`
	Name             string `json:"name"`
}

type patchGroupEventItem struct {
	EventID    string `json:"event_id"`
	StartTime  string `json:"start_time"`
	GameModeID string `json:"game_mode_id"`
}

type updateEventGroupSettingsRequest struct {
	Region           string                `json:"region"`
	SubMin           int32                 `json:"sub_min"`
	SortLogic        string                `json:"sort_logic"`
	RegistrationOpen *bool                 `json:"registration_open"`
	Name             string                `json:"name"`
	Events           []patchGroupEventItem `json:"events"`
}

type updateRegistrationStatusRequest struct {
	RegistrationOpen bool `json:"registration_open"`
}

type upsertRegistrationRequest struct {
	CanSubstitute bool   `json:"can_substitute"`
	CanLobbyHost  bool   `json:"can_lobby_host"`
	DuoRequest    string `json:"duo_request"`
}

type bulkRegistrationEventRequest struct {
	EventID       string `json:"event_id"`
	CanSubstitute bool   `json:"can_substitute"`
	CanLobbyHost  bool   `json:"can_lobby_host"`
}

type upsertGroupRegistrationRequest struct {
	DuoRequest string                         `json:"duo_request"`
	Events     []bulkRegistrationEventRequest `json:"events"`
}

type swapPlayersRequest struct {
	UserIDA string `json:"user_id_a"`
	UserIDB string `json:"user_id_b"`
}

type setLobbyHostRequest struct {
	UserID string `json:"user_id"`
}

type movePlacementRequest struct {
	UserID string `json:"user_id"`
}

type moveUnplacedToSubsRequest struct {
	UserID  string `json:"user_id"`
	LobbyID string `json:"lobby_id"`
}

// POST /events
func (h *Handler) CreateEventHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	var body createEventRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid request body in CreateEventHandler", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Improper json or json value types",
		})
		return
	}

	gameModeID, err := uuid.Parse(body.GameModeID)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid game_mode_id in CreateEventHandler", "user_id", userUUID, "game_mode_id", body.GameModeID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "game_mode_id must be a valid UUID",
		})
		return
	}

	region := strings.TrimSpace(body.Region)
	if region == "" {
		slog.WarnContext(c.Request.Context(), "missing region in CreateEventHandler", "user_id", userUUID)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "region is required",
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, body.StartTime)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid start_time in CreateEventHandler", "user_id", userUUID, "start_time", body.StartTime, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_time must be RFC3339 datetime",
		})
		return
	}

	if startTime.Before(time.Now()) {
		slog.WarnContext(c.Request.Context(), "start_time in past in CreateEventHandler", "user_id", userUUID, "start_time", startTime)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_time cannot be in the past",
		})
		return
	}

	if body.SubMin < 0 {
		slog.WarnContext(c.Request.Context(), "invalid sub_min in CreateEventHandler", "user_id", userUUID, "sub_min", body.SubMin)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "sub_min cannot be negative",
		})
		return
	}

	if body.GamesToRun <= 0 {
		slog.WarnContext(c.Request.Context(), "invalid games_to_run in CreateEventHandler", "user_id", userUUID, "games_to_run", body.GamesToRun)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "games_to_run must be greater than 0",
		})
		return
	}

	sortLogic := strings.TrimSpace(body.SortLogic)
	if sortLogic == "" {
		sortLogic = "balanced"
	}
	if sortLogic != "balanced" && sortLogic != "ranked" {
		slog.WarnContext(c.Request.Context(), "invalid sort_logic in CreateEventHandler", "user_id", userUUID, "sort_logic", body.SortLogic)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "sort_logic must be 'balanced' or 'ranked'",
		})
		return
	}

	name, err := textinput.NormalizeOptional(body.Name, eventGroupNameMaxRunes)
	if err != nil {
		if errors.Is(err, textinput.ErrTooLong) {
			slog.WarnContext(c.Request.Context(), "name too long in CreateEventHandler", "user_id", userUUID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "name must be at most 50 characters",
			})
			return
		}
		if errors.Is(err, textinput.ErrInvalidChars) {
			slog.WarnContext(c.Request.Context(), "name has invalid characters in CreateEventHandler", "user_id", userUUID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "name contains invalid characters",
			})
			return
		}
		slog.WarnContext(c.Request.Context(), "invalid name in CreateEventHandler", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "name is invalid",
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
		sortLogic,
		name,
		startTime,
		body.GamesToRun,
	)
	if err != nil {
		if errors.Is(err, store.ErrInvalidSortLogic) {
			slog.WarnContext(c.Request.Context(), "invalid sort_logic from store in CreateEventHandler", "user_id", userUUID, "sort_logic", sortLogic)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "sort_logic must be 'balanced' or 'ranked'",
			})
			return
		}
		if errors.Is(err, store.ErrGameModeNotFound) || errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(c.Request.Context(), "game mode not found in CreateEventHandler", "user_id", userUUID, "game_mode_id", gameModeID)
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
		if errors.Is(err, store.ErrEventGroupNotFound) || errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(c.Request.Context(), "event group not found", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "failed to fetch event group detail", "group_id", groupID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch event group"})
		return
	}

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
	if body.RegistrationOpen == nil {
		slog.WarnContext(c.Request.Context(), "missing registration_open in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", groupID)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "registration_open is required"})
		return
	}

	sortLogicInput := strings.TrimSpace(body.SortLogic)
	if sortLogicInput != "" && sortLogicInput != "balanced" && sortLogicInput != "ranked" {
		slog.WarnContext(c.Request.Context(), "invalid sort_logic in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", groupID, "sort_logic", body.SortLogic)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "sort_logic must be 'balanced' or 'ranked'",
		})
		return
	}

	if len(body.Events) == 0 {
		slog.WarnContext(c.Request.Context(), "missing events in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", groupID)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "events must include every scheduled game in this group"})
		return
	}

	eventUpdates := make([]store.GroupEventUpdate, 0, len(body.Events))
	for _, item := range body.Events {
		eventID, err := uuid.Parse(strings.TrimSpace(item.EventID))
		if err != nil || eventID == uuid.Nil {
			slog.WarnContext(c.Request.Context(), "invalid event_id in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", groupID, "event_id", item.EventID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "each events[].event_id must be a valid UUID"})
			return
		}
		gameModeID, err := uuid.Parse(strings.TrimSpace(item.GameModeID))
		if err != nil || gameModeID == uuid.Nil {
			slog.WarnContext(c.Request.Context(), "invalid game_mode_id in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", groupID, "event_id", eventID, "game_mode_id", item.GameModeID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "each events[].game_mode_id must be a valid UUID"})
			return
		}
		startTime, err := time.Parse(time.RFC3339, item.StartTime)
		if err != nil {
			slog.WarnContext(c.Request.Context(), "invalid start_time in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", groupID, "event_id", eventID, "start_time", item.StartTime, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "each events[].start_time must be RFC3339 datetime"})
			return
		}
		eventUpdates = append(eventUpdates, store.GroupEventUpdate{
			EventID:    eventID,
			StartTime:  startTime,
			GameModeID: gameModeID,
		})
	}

	name, err := textinput.NormalizeOptional(body.Name, eventGroupNameMaxRunes)
	if err != nil {
		if errors.Is(err, textinput.ErrTooLong) {
			slog.WarnContext(c.Request.Context(), "name too long in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "name must be at most 50 characters",
			})
			return
		}
		if errors.Is(err, textinput.ErrInvalidChars) {
			slog.WarnContext(c.Request.Context(), "name has invalid characters in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "name contains invalid characters",
			})
			return
		}
		slog.WarnContext(c.Request.Context(), "invalid name in UpdateEventGroupSettingsHandler", "user_id", userUUID, "group_id", groupID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "name is invalid",
		})
		return
	}

	err = h.store.UpdateEventGroupSettings(c.Request.Context(), groupID, userUUID, body.Region, body.SubMin, sortLogicInput, *body.RegistrationOpen, name, eventUpdates)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden event group settings update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can edit this event group"})
		case errors.Is(err, store.ErrInvalidSortLogic):
			slog.WarnContext(c.Request.Context(), "invalid sort_logic for event group settings update", "user_id", userUUID, "group_id", groupID, "sort_logic", body.SortLogic)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "sort_logic must be 'balanced' or 'ranked'"})
		case errors.Is(err, store.ErrInvalidSubMin):
			slog.WarnContext(c.Request.Context(), "invalid event group settings payload", "user_id", userUUID, "group_id", groupID, "region", body.Region, "sub_min", body.SubMin)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "region is required and sub_min must be >= 0"})
		case errors.Is(err, store.ErrOpenRegistrationTeams):
			slog.WarnContext(c.Request.Context(), "registration open blocked due to existing teams in settings update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Delete teams before opening registration"})
		case errors.Is(err, store.ErrGameModeNotFound):
			slog.WarnContext(c.Request.Context(), "game mode not found for settings update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Game mode not found"})
		case errors.Is(err, store.ErrGameModeWrongGame):
			slog.WarnContext(c.Request.Context(), "game mode wrong game for settings update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Game mode must belong to this event's game"})
		case errors.Is(err, store.ErrInvalidGroupEvents):
			slog.WarnContext(c.Request.Context(), "invalid events payload for settings update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "events must include each scheduled game exactly once"})
		case errors.Is(err, store.ErrEventStartInPast):
			slog.WarnContext(c.Request.Context(), "event start in past for settings update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "start_time cannot be in the past"})
		case errors.Is(err, store.ErrEventGroupNotFound), errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for settings update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to update event group settings", "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update event group"})
		}
		return
	}

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
		case errors.Is(err, store.ErrEventGroupNotFound), errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for delete", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to delete event group", "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to delete event group"})
		}
		return
	}

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
		case errors.Is(err, store.ErrEventGroupNotFound), errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for registration status update", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to update registration status", "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update registration status"})
		}
		return
	}

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

	err = h.store.CreateTeamsForGroup(c.Request.Context(), groupID, userUUID, h.matchmakingSettings)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInsufficientPlayers), errors.Is(err, store.ErrInsufficientSubstitutes):
			msg := err.Error()
			var teamErr *store.TeamCreationError
			if errors.As(err, &teamErr) {
				msg = teamErr.Message
			}
			slog.WarnContext(c.Request.Context(), "create teams validation failed", "user_id", userUUID, "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": msg})
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden create teams attempt", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can create teams"})
		case errors.Is(err, store.ErrTeamsAlreadyCreated):
			slog.WarnContext(c.Request.Context(), "create teams blocked because teams already exist", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Teams already exist for this event group"})
		case errors.Is(err, store.ErrEventGroupNotFound), errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for create teams", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to create teams", "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create teams"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// DELETE /events/:groupId/teams
func (h *Handler) DeleteTeamsHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid groupId in DeleteTeamsHandler", "user_id", userUUID, "group_id", c.Param("groupId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "groupId must be a valid UUID"})
		return
	}

	err = h.store.DeleteTeamsForGroup(c.Request.Context(), groupID, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden delete teams attempt", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can delete teams"})
		case errors.Is(err, store.ErrTeamsNotCreated):
			slog.WarnContext(c.Request.Context(), "delete teams requested but no teams exist", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "No teams exist for this event group"})
		case errors.Is(err, store.ErrEventGroupNotFound), errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for delete teams", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to delete teams", "group_id", groupID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to delete teams"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /registrations/:eventId/player-swap
// Host-only. Atomically exchanges two players' placements for one game.
func (h *Handler) SwapPlayersHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid eventId in SwapPlayersHandler", "user_id", userUUID, "event_id", c.Param("eventId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "eventId must be a valid UUID"})
		return
	}

	var body swapPlayersRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid request body in SwapPlayersHandler", "user_id", userUUID, "event_id", eventID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Improper json or json value types"})
		return
	}

	userA, err := uuid.Parse(body.UserIDA)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid user_id_a in SwapPlayersHandler", "user_id", userUUID, "event_id", eventID, "user_id_a", body.UserIDA, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "user_id_a must be a valid UUID"})
		return
	}
	userB, err := uuid.Parse(body.UserIDB)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid user_id_b in SwapPlayersHandler", "user_id", userUUID, "event_id", eventID, "user_id_b", body.UserIDB, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "user_id_b must be a valid UUID"})
		return
	}

	err = h.store.SwapPlayersForEvent(c.Request.Context(), eventID, userUUID, userA, userB, h.matchmakingSettings)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden swap players attempt", "user_id", userUUID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can swap players"})
		case errors.Is(err, store.ErrInsufficientSubstitutes):
			msg := err.Error()
			var teamErr *store.TeamCreationError
			if errors.As(err, &teamErr) {
				msg = teamErr.Message
			}
			slog.WarnContext(c.Request.Context(), "swap players validation failed", "user_id", userUUID, "event_id", eventID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": msg})
		case errors.Is(err, store.ErrInvalidPlayerSwap):
			msg := err.Error()
			var swapErr *store.SwapValidationError
			if errors.As(err, &swapErr) {
				msg = swapErr.Message
			}
			slog.WarnContext(c.Request.Context(), "swap players validation failed", "user_id", userUUID, "event_id", eventID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": msg})
		case errors.Is(err, store.ErrTeamsNotCreated):
			slog.WarnContext(c.Request.Context(), "swap players requested but no teams exist", "user_id", userUUID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "No teams exist for this game"})
		case errors.Is(err, store.ErrEventNotFound), errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event not found for swap players", "user_id", userUUID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to swap players", "event_id", eventID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to swap players"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) respondPlacementMoveError(c *gin.Context, userUUID, eventID uuid.UUID, err error, action string) {
	switch {
	case errors.Is(err, store.ErrForbidden):
		slog.WarnContext(c.Request.Context(), "forbidden placement move attempt", "user_id", userUUID, "event_id", eventID, "action", action)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can move players"})
	case errors.Is(err, store.ErrInsufficientSubstitutes):
		msg := err.Error()
		var teamErr *store.TeamCreationError
		if errors.As(err, &teamErr) {
			msg = teamErr.Message
		}
		slog.WarnContext(c.Request.Context(), "placement move validation failed", "user_id", userUUID, "event_id", eventID, "action", action, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": msg})
	case errors.Is(err, store.ErrInvalidPlayerSwap):
		msg := err.Error()
		var swapErr *store.SwapValidationError
		if errors.As(err, &swapErr) {
			msg = swapErr.Message
		}
		slog.WarnContext(c.Request.Context(), "placement move validation failed", "user_id", userUUID, "event_id", eventID, "action", action, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": msg})
	case errors.Is(err, store.ErrTeamsNotCreated):
		slog.WarnContext(c.Request.Context(), "placement move requested but no teams exist", "user_id", userUUID, "event_id", eventID, "action", action)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "No teams exist for this game"})
	case errors.Is(err, store.ErrEventNotFound), errors.Is(err, pgx.ErrNoRows):
		slog.WarnContext(c.Request.Context(), "event not found for placement move", "user_id", userUUID, "event_id", eventID, "action", action)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event not found"})
	default:
		slog.ErrorContext(c.Request.Context(), "failed placement move", "event_id", eventID, "action", action, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to move player"})
	}
}

// POST /registrations/:eventId/sub-to-unplaced
// Host-only. Removes a substitute from their lobby sub pool.
func (h *Handler) MoveSubToUnplacedHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid eventId in MoveSubToUnplacedHandler", "user_id", userUUID, "event_id", c.Param("eventId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "eventId must be a valid UUID"})
		return
	}

	var body movePlacementRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid request body in MoveSubToUnplacedHandler", "user_id", userUUID, "event_id", eventID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Improper json or json value types"})
		return
	}

	targetUserID, err := uuid.Parse(body.UserID)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid user_id in MoveSubToUnplacedHandler", "user_id", userUUID, "event_id", eventID, "target_user_id", body.UserID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "user_id must be a valid UUID"})
		return
	}

	err = h.store.MoveSubToUnplacedForEvent(c.Request.Context(), eventID, userUUID, targetUserID, h.matchmakingSettings)
	if err != nil {
		h.respondPlacementMoveError(c, userUUID, eventID, err, "sub-to-unplaced")
		return
	}

	c.Status(http.StatusNoContent)
}

// POST /registrations/:eventId/unplaced-to-subs
// Host-only. Adds an unplaced substitute-eligible player to a lobby sub pool.
func (h *Handler) MoveUnplacedToSubsHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid eventId in MoveUnplacedToSubsHandler", "user_id", userUUID, "event_id", c.Param("eventId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "eventId must be a valid UUID"})
		return
	}

	var body moveUnplacedToSubsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid request body in MoveUnplacedToSubsHandler", "user_id", userUUID, "event_id", eventID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Improper json or json value types"})
		return
	}

	targetUserID, err := uuid.Parse(body.UserID)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid user_id in MoveUnplacedToSubsHandler", "user_id", userUUID, "event_id", eventID, "target_user_id", body.UserID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "user_id must be a valid UUID"})
		return
	}
	lobbyID, err := uuid.Parse(body.LobbyID)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid lobby_id in MoveUnplacedToSubsHandler", "user_id", userUUID, "event_id", eventID, "lobby_id", body.LobbyID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "lobby_id must be a valid UUID"})
		return
	}

	err = h.store.MoveUnplacedToSubsForEvent(c.Request.Context(), eventID, userUUID, targetUserID, lobbyID, h.matchmakingSettings)
	if err != nil {
		h.respondPlacementMoveError(c, userUUID, eventID, err, "unplaced-to-subs")
		return
	}

	c.Status(http.StatusNoContent)
}

// POST /registrations/:eventId/lobby-host
// Host-only. Assigns a team player as the lobby host for their lobby.
func (h *Handler) SetLobbyHostHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid eventId in SetLobbyHostHandler", "user_id", userUUID, "event_id", c.Param("eventId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "eventId must be a valid UUID"})
		return
	}

	var body setLobbyHostRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid request body in SetLobbyHostHandler", "user_id", userUUID, "event_id", eventID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Improper json or json value types"})
		return
	}

	targetUserID, err := uuid.Parse(body.UserID)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid user_id in SetLobbyHostHandler", "user_id", userUUID, "event_id", eventID, "target_user_id", body.UserID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "user_id must be a valid UUID"})
		return
	}

	err = h.store.SetLobbyHostForEvent(c.Request.Context(), eventID, userUUID, targetUserID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrForbidden):
			slog.WarnContext(c.Request.Context(), "forbidden set lobby host attempt", "user_id", userUUID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only the host can change the lobby host"})
		case errors.Is(err, store.ErrInvalidLobbyHostChange):
			msg := err.Error()
			var hostErr *store.LobbyHostValidationError
			if errors.As(err, &hostErr) {
				msg = hostErr.Message
			}
			slog.WarnContext(c.Request.Context(), "set lobby host validation failed", "user_id", userUUID, "event_id", eventID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": msg})
		case errors.Is(err, store.ErrTeamsNotCreated):
			slog.WarnContext(c.Request.Context(), "set lobby host requested but no teams exist", "user_id", userUUID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "No teams exist for this game"})
		case errors.Is(err, store.ErrEventNotFound), errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event not found for set lobby host", "user_id", userUUID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to set lobby host", "event_id", eventID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to set lobby host"})
		}
		return
	}

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
		case errors.Is(err, store.ErrUserGameProfileIncomplete):
			slog.WarnContext(c.Request.Context(), "registration upsert rejected due to incomplete user game profile", "user_id", userUUID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Complete your game profile before registering"})
		case errors.Is(err, store.ErrEventNotFound), errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event not found for registration upsert", "user_id", userUUID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to upsert registration", "event_id", eventID, "user_id", userUUID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to save registration"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// PUT /registrations/group/:groupId/me
func (h *Handler) UpsertMyGroupRegistrationsHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}
	groupID, err := uuid.Parse(c.Param("groupId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid groupId in UpsertMyGroupRegistrationsHandler", "user_id", userUUID, "group_id", c.Param("groupId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "groupId must be a valid UUID"})
		return
	}

	var body upsertGroupRegistrationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid request body in UpsertMyGroupRegistrationsHandler", "user_id", userUUID, "group_id", groupID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Improper json or json value types"})
		return
	}
	if len(body.Events) == 0 {
		slog.WarnContext(c.Request.Context(), "group registration upsert rejected due to empty event selection", "user_id", userUUID, "group_id", groupID)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "At least one event must be selected"})
		return
	}

	items := make([]store.RegistrationUpsertItem, 0, len(body.Events))
	seenEventIDs := make(map[uuid.UUID]struct{}, len(body.Events))
	for _, requestedEvent := range body.Events {
		eventID, err := uuid.Parse(requestedEvent.EventID)
		if err != nil {
			slog.WarnContext(c.Request.Context(), "invalid event_id in UpsertMyGroupRegistrationsHandler", "user_id", userUUID, "group_id", groupID, "event_id", requestedEvent.EventID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "event_id must be a valid UUID"})
			return
		}
		if _, exists := seenEventIDs[eventID]; exists {
			slog.WarnContext(c.Request.Context(), "duplicate event_id in UpsertMyGroupRegistrationsHandler", "user_id", userUUID, "group_id", groupID, "event_id", eventID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Duplicate event_id values are not allowed"})
			return
		}
		seenEventIDs[eventID] = struct{}{}

		items = append(items, store.RegistrationUpsertItem{
			EventID:       eventID,
			CanSubstitute: requestedEvent.CanSubstitute,
			CanLobbyHost:  requestedEvent.CanLobbyHost,
		})
	}

	var duoRequest *string
	if trimmed := strings.TrimSpace(body.DuoRequest); trimmed != "" {
		duoRequest = &trimmed
	}

	err = h.store.UpsertRegistrationsForGroup(c.Request.Context(), groupID, userUUID, items, duoRequest)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidRegistration):
			slog.WarnContext(c.Request.Context(), "group registration upsert rejected due to invalid registration payload", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "At least one event must be selected"})
		case errors.Is(err, store.ErrRegistrationClosed):
			slog.WarnContext(c.Request.Context(), "group registration upsert rejected because registration is closed", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Registration is closed"})
		case errors.Is(err, store.ErrRegistrationDeleteWithTeams):
			slog.WarnContext(c.Request.Context(), "group registration upsert rejected because teams exist", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Cannot delete registration after teams have been created"})
		case errors.Is(err, store.ErrUserGameProfileIncomplete):
			slog.WarnContext(c.Request.Context(), "group registration upsert rejected due to incomplete user game profile", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Complete your game profile before registering"})
		case errors.Is(err, store.ErrEventNotFound):
			slog.WarnContext(c.Request.Context(), "group registration upsert rejected due to invalid events for group", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "One or more selected events are invalid for this group"})
		case errors.Is(err, store.ErrEventGroupNotFound), errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event group not found for group registration upsert", "user_id", userUUID, "group_id", groupID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to upsert group registrations", "group_id", groupID, "user_id", userUUID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to save registrations"})
		}
		return
	}

	slog.InfoContext(c.Request.Context(), "upserted group registrations", "group_id", groupID, "user_id", userUUID, "event_count", len(items))
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
		case errors.Is(err, store.ErrRegistrationDeleteWithTeams):
			slog.WarnContext(c.Request.Context(), "registration delete rejected because teams exist", "actor_user_id", actorID, "event_id", eventID, "target_user_id", targetUserID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Cannot delete registration after teams have been created"})
		case errors.Is(err, store.ErrRegistrationNotFound):
			slog.WarnContext(c.Request.Context(), "registration not found for delete", "actor_user_id", actorID, "event_id", eventID, "target_user_id", targetUserID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Registration not found"})
		case errors.Is(err, store.ErrEventNotFound), errors.Is(err, pgx.ErrNoRows):
			slog.WarnContext(c.Request.Context(), "event not found for registration delete", "actor_user_id", actorID, "event_id", eventID, "target_user_id", targetUserID)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "failed to delete registration", "event_id", eventID, "target_user_id", targetUserID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to delete registration"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
