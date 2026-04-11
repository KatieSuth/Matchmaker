package handler

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /users/me
func (h *Handler) UsersMeHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached GetSystemGamesHandler without userID in context")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		log.Printf("Error parsing user ID string (%s) into UUID: %v", userID, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not parse user ID",
		})
		return
	}

	user, err := h.store.GetUserByUserID(c.Request.Context(), userUUID)
	if err != nil {
		log.Printf("Error fetching user (%s): %v", userUUID.String(), err)
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
	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached GetSystemGamesHandler without userID in context")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to parse user ID into UUID", "user_id", userID, "error", err)
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

	err = c.ShouldBindJSON(&body)

	if err != nil {
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
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": err.Error(),
				})
				return err
			} else if err != nil {
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
	userID, exists := c.Get("userID")
	if !exists {
		slog.WarnContext(c.Request.Context(), "request reached GetSystemGamesHandler without userID in context")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		log.Printf("Error parsing user ID string (%s) into UUID: %v", userID, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Could not parse user ID",
		})
		return
	}

	userGames, err := h.store.GetUserGamesForUser(c.Request.Context(), userUUID)
	if err != nil {
		log.Printf("Error fetching user (%s): %v", userUUID.String(), err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to fetch user",
		})
		return
	}

	c.JSON(http.StatusOK, userGames)
}
