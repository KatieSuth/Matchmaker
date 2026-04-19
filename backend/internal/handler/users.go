package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /users/me
func (h *Handler) UsersMeHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached UsersMeHandler without userID in context")
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

	user, err := h.store.GetUserByUserID(c.Request.Context(), userUUID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to fetch user", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch user",
		})
		return
	}

	slog.DebugContext(c.Request.Context(), "user fetched successfully", "user_id", userUUID)
	c.JSON(http.StatusOK, user)
}

// PUT /users/me
func (h *Handler) UpdateUsersMeHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached UpdateUsersMeHandler without userID in context")
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

	var body struct {
		Pronouns     string           `json:"pronouns"`
		ShowPronouns bool             `json:"show_pronouns"`
		Region       string           `json:"region"`
		Games        []model.UserGame `json:"games"`
	}

	if err = c.ShouldBindJSON(&body); err != nil {
		slog.WarnContext(c.Request.Context(), "failed to bind request body", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Improper json or json value types",
		})
		return
	}

	type result struct {
		model.User
		Games []model.UserGame `json:"games"`
	}

	var res result

	err = h.store.WithTx(c.Request.Context(), func(tx store.Store) error {
		user, err := h.store.UpdateUser(c.Request.Context(), userUUID, &body.Pronouns, body.ShowPronouns, &body.Region)
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
		res.ID = user.ID
		res.ImageUrl = user.ImageUrl
		res.NewUser = user.NewUser
		res.Pronouns = user.Pronouns
		res.ShowPronouns = user.ShowPronouns
		res.UpdatedAt = user.UpdatedAt

		var games []model.UserGame
		for _, ug := range body.Games {
			userGame, result, err := h.store.UpsertGameForUser(c.Request.Context(), userUUID, ug)
			if result == http.StatusBadRequest {
				slog.WarnContext(c.Request.Context(), "bad request upserting game for user", "user_id", userUUID, "game", ug, "error", err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": err.Error(),
				})
				return err
			} else if err != nil {
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

	slog.InfoContext(c.Request.Context(), "user updated successfully", "user_id", userUUID, "game_count", len(res.Games))
	c.JSON(http.StatusOK, res)
}

// GET /users/me/games
func (h *Handler) UsersMeGamesHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached UsersMeGamesHandler without userID in context")
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

	userGames, err := h.store.GetUserGamesForUser(c.Request.Context(), userUUID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to fetch user games", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch user",
		})
		return
	}

	slog.DebugContext(c.Request.Context(), "user games fetched successfully", "user_id", userUUID, "count", len(userGames))
	c.JSON(http.StatusOK, userGames)
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
	}

	type Response struct {
		Events     []model.DashboardEvent `json:"events"`
		NextCursor string                 `json:"next_cursor"`
		HasMore    bool                   `json:"has_more"`
	}

	var params QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid query parameters in UsersMeEventsHandler")

	}

	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached UsersMeEventsHandler without userID in context")
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

	hosting := params.Hosting
	past := params.Past
	from := params.From
	to := params.To
	gameId := params.GameId
	cursor := params.Cursor

	var dateFrom *time.Time
	var dateTo *time.Time
	if from != "" {
		if t, err := time.Parse("2026-01-31", from); err == nil {
			dateFrom = &t
		} else {
			slog.WarnContext(c.Request.Context(), "invalid 'from' format", "user_id", userID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "invalid 'from' format",
			})
		}
	}

	if to != "" {
		if t, err := time.Parse("2026-01-31", to); err == nil {
			dateTo = &t
		} else {
			slog.WarnContext(c.Request.Context(), "invalid 'to' format", "user_id", userID, "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "invalid 'to' format",
			})
		}
	}

	events, hasMore, nextCursor, err := h.store.GetEventsForUser(c.Request.Context(), userUUID, hosting, past, dateFrom, dateTo, gameId, cursor)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to fetch user's events", "user_id", userUUID, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch events",
		})
		return
	}

	response := Response{
		Events:     events,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}

	slog.DebugContext(c.Request.Context(), "user events fetched successfully", "user_id", userUUID, "count", len(events))
	c.JSON(http.StatusOK, response)
}
