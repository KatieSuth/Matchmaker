package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/textinput"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const userDisplayNameMaxRunes = 50

type updateUsersMeRequest struct {
	DisplayName  string           `json:"display_name"`
	Pronouns     string           `json:"pronouns"`
	ShowPronouns bool             `json:"show_pronouns"`
	Region       string           `json:"region"`
	Games        []model.UserGame `json:"games"`
}

// GET /users/me
func (h *Handler) UsersMeHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	user, err := h.store.GetUserByUserID(c.Request.Context(), userUUID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to fetch user", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch user",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// PUT /users/me
func (h *Handler) UpdateUsersMeHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	var body updateUsersMeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "failed to bind request body", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Improper json or json value types",
		})
		return
	}

	displayName, err := textinput.NormalizeOptional(body.DisplayName, userDisplayNameMaxRunes)
	if err != nil {
		if errors.Is(err, textinput.ErrTooLong) {
			slog.WarnContext(c.Request.Context(), "display_name too long in UpdateUsersMeHandler", "user_id", userUUID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "display_name must be at most 50 characters",
			})
			return
		}
		if errors.Is(err, textinput.ErrInvalidChars) {
			slog.WarnContext(c.Request.Context(), "display_name has invalid characters in UpdateUsersMeHandler", "user_id", userUUID)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "display_name contains invalid characters",
			})
			return
		}
		slog.WarnContext(c.Request.Context(), "invalid display_name in UpdateUsersMeHandler", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "display_name is invalid",
		})
		return
	}
	var displayNamePtr *string
	if displayName != "" {
		displayNamePtr = &displayName
	}

	type result struct {
		model.User
		Games []model.UserGame `json:"games"`
	}

	var res result

	err = h.store.WithTx(c.Request.Context(), func(tx store.Store) error {
		user, err := tx.UpdateUser(c.Request.Context(), userUUID, displayNamePtr, &body.Pronouns, body.ShowPronouns, &body.Region)
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "failed to update user", "user_id", userUUID, "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Something went wrong",
			})
			return err
		}

		res.CreatedAt = user.CreatedAt
		res.DiscordID = user.DiscordID
		res.DiscordName = user.DiscordName
		res.DisplayName = user.DisplayName
		res.ID = user.ID
		res.ImageUrl = user.ImageUrl
		res.NewUser = user.NewUser
		res.Pronouns = user.Pronouns
		res.ShowPronouns = user.ShowPronouns
		res.Region = user.Region
		res.UpdatedAt = user.UpdatedAt

		var games []model.UserGame
		for _, ug := range body.Games {
			userGame, err := tx.UpsertGameForUser(c.Request.Context(), userUUID, ug)
			if errors.Is(err, store.ErrInvalidGame) ||
				errors.Is(err, store.ErrCurrentRankMissing) ||
				errors.Is(err, store.ErrPeakRankMissing) ||
				errors.Is(err, store.ErrInvalidCurrentRank) ||
				errors.Is(err, store.ErrInvalidPeakRank) ||
				errors.Is(err, store.ErrInvalidRankOrder) {
				slog.WarnContext(c.Request.Context(), "bad request upserting game for user", "user_id", userUUID, "game", ug, "error", err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": err.Error(),
				})
				return err
			}
			if err != nil {
				slog.ErrorContext(c.Request.Context(), "failed to upsert game for user", "user_id", userUUID, "game", ug, "error", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "Something went wrong",
				})
				return err
			}
			games = append(games, userGame)
		}

		res.Games = games
		return nil
	})

	if err != nil {
		return
	}

	c.JSON(http.StatusOK, res)
}

// GET /users/me/games
func (h *Handler) UsersMeGamesHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	userGames, err := h.store.GetUserGamesForUser(c.Request.Context(), userUUID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to fetch user games", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch user",
		})
		return
	}

	c.JSON(http.StatusOK, userGames)
}

// PUT /users/me/games/:gameId
func (h *Handler) UpsertUsersMeGameHandler(c *gin.Context) {
	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	gameID, err := uuid.Parse(c.Param("gameId"))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid gameId in UpsertUsersMeGameHandler", "user_id", userUUID, "game_id", c.Param("gameId"), "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "gameId must be a valid UUID",
		})
		return
	}

	var body struct {
		InGameName  string `json:"in_game_name"`
		CurrentRank string `json:"current_rank"`
		PeakRank    string `json:"peak_rank"`
		ShowRank    bool   `json:"show_rank"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "failed to bind game upsert body", "user_id", userUUID, "game_id", gameID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Improper json or json value types",
		})
		return
	}

	currentRankID, err := uuid.Parse(strings.TrimSpace(body.CurrentRank))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid current_rank in UpsertUsersMeGameHandler", "user_id", userUUID, "game_id", gameID, "current_rank", body.CurrentRank, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "current_rank must be a valid UUID",
		})
		return
	}
	peakRankID, err := uuid.Parse(strings.TrimSpace(body.PeakRank))
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid peak_rank in UpsertUsersMeGameHandler", "user_id", userUUID, "game_id", gameID, "peak_rank", body.PeakRank, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "peak_rank must be a valid UUID",
		})
		return
	}
	inGameName := strings.TrimSpace(body.InGameName)
	if inGameName == "" {
		slog.WarnContext(c.Request.Context(), "missing in_game_name in UpsertUsersMeGameHandler", "user_id", userUUID, "game_id", gameID)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "in_game_name is required",
		})
		return
	}

	_, err = h.store.UpsertGameForUser(c.Request.Context(), userUUID, model.UserGame{
		GameID:      gameID,
		InGameName:  &inGameName,
		CurrentRank: &currentRankID,
		PeakRank:    &peakRankID,
		ShowRank:    body.ShowRank,
	})
	if err != nil {
		if errors.Is(err, store.ErrInvalidGame) ||
			errors.Is(err, store.ErrCurrentRankMissing) ||
			errors.Is(err, store.ErrPeakRankMissing) ||
			errors.Is(err, store.ErrInvalidCurrentRank) ||
			errors.Is(err, store.ErrInvalidPeakRank) ||
			errors.Is(err, store.ErrInvalidRankOrder) {
			slog.WarnContext(c.Request.Context(), "invalid game upsert payload", "user_id", userUUID, "game_id", gameID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		slog.ErrorContext(c.Request.Context(), "failed to upsert game for user", "user_id", userUUID, "game_id", gameID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Something went wrong",
		})
		return
	}

	slog.InfoContext(c.Request.Context(), "upserted user game", "user_id", userUUID, "game_id", gameID)
	c.Status(http.StatusNoContent)
}

// GET /users/me/events
func (h *Handler) UsersMeEventsHandler(c *gin.Context) {
	type QueryParams struct {
		Hosting bool   `form:"hosting"`
		Past    bool   `form:"past"`
		From    string `form:"from"`
		To      string `form:"to"`
		GameId  string `form:"game_id"`
		Cursor  string `form:"cursor"`
		Tz      string `form:"tz"`
	}

	type Response struct {
		EventGroups []model.DashboardEvent `json:"event_groups"`
		NextCursor  string                 `json:"next_cursor"`
		HasMore     bool                   `json:"has_more"`
	}

	var params QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid query parameters in UsersMeEventsHandler")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid query parameters",
		})
		return
	}

	userUUID, ok := userIDFromContext(c)
	if !ok {
		return
	}

	hosting := params.Hosting
	past := params.Past
	from := params.From
	to := params.To
	gameId := params.GameId
	cursor := params.Cursor
	timezone := params.Tz

	if timezone == "" {
		slog.WarnContext(c.Request.Context(), "missing timezone query parameter", "user_id", userUUID)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "tz is required",
		})
		return
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "invalid timezone", "user_id", userUUID, "tz", timezone, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid timezone",
		})
		return
	}

	var dateFrom *time.Time
	var dateTo *time.Time
	if from != "" {
		if t, err := time.ParseInLocation("2006-01-02", from, location); err == nil {
			dateFrom = &t
		} else {
			slog.WarnContext(c.Request.Context(), "invalid 'from' format", "user_id", userUUID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "invalid 'from' format",
			})
			return
		}
	}

	if to != "" {
		if t, err := time.ParseInLocation("2006-01-02", to, location); err == nil {
			dateTo = &t
		} else {
			slog.WarnContext(c.Request.Context(), "invalid 'to' format", "user_id", userUUID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "invalid 'to' format",
			})
			return
		}
	}

	if dateFrom != nil && dateTo != nil && dateFrom.After(*dateTo) {
		slog.WarnContext(c.Request.Context(), "invalid date range in UsersMeEventsHandler", "user_id", userUUID, "from", from, "to", to)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "'from' must be before or equal to 'to'",
		})
		return
	}

	eventGroups, hasMore, nextCursor, err := h.store.GetEventsForUser(c.Request.Context(), userUUID, hosting, past, dateFrom, dateTo, gameId, cursor, timezone)
	if err != nil {
		if errors.Is(err, store.ErrInvalidGameID) || errors.Is(err, store.ErrInvalidCursor) || errors.Is(err, store.ErrInvalidTimezone) {
			slog.WarnContext(c.Request.Context(), "invalid users/me/events query values", "user_id", userUUID, "game_id", gameId, "cursor", cursor, "tz", timezone, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}

		slog.ErrorContext(c.Request.Context(), "failed to fetch user's events", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch events",
		})
		return
	}

	response := Response{
		EventGroups: eventGroups,
		HasMore:     hasMore,
		NextCursor:  nextCursor,
	}

	c.JSON(http.StatusOK, response)
}
