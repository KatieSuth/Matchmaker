package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/KatieSuth/MatchmakerAPI/internal/discord"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrDiscordGuildRestricted means the viewer is not in a required Discord server (expected 403).
var ErrDiscordGuildRestricted = errors.New("discord guild restricted")

// errDiscordGuildNotMember means the host selected a server they do not belong to.
var errDiscordGuildNotMember = errors.New("host is not a member of a selected discord guild")

// discordGuildRestriction carries 403 details for a failed Discord server lock check.
type discordGuildRestriction struct {
	eventTitle string
	eventNamed bool // true when eventTitle is a host-assigned name, not the game name
	guilds     []model.DiscordGuild
	cause      error // nil = not a member; non-nil = Discord/token failure during the check
}

func (e *discordGuildRestriction) Error() string {
	if e.cause != nil {
		return e.cause.Error()
	}
	return ErrDiscordGuildRestricted.Error()
}

func (e *discordGuildRestriction) Unwrap() error {
	if e.cause != nil {
		return e.cause
	}
	return ErrDiscordGuildRestricted
}

// isEventDiscordAllowed reports whether userID may open or mutate a group under its Discord lock.
// Owners and unrestricted groups (no junction rows) are allowed without calling Discord.
func (h *Handler) isEventDiscordAllowed(ctx context.Context, groupID, userID uuid.UUID) error {
	ownerID, title, named, err := h.store.GetEventGroupAccessMeta(ctx, groupID)
	if err != nil {
		return err
	}
	if userID == ownerID {
		return nil
	}
	guilds, err := h.store.ListEventGroupDiscordGuilds(ctx, groupID)
	if err != nil {
		return err
	}
	if len(guilds) == 0 {
		return nil
	}
	required := make([]string, 0, len(guilds))
	for _, g := range guilds {
		required = append(required, g.ID)
	}
	if h.discord == nil {
		return &discordGuildRestriction{eventTitle: title, eventNamed: named, guilds: guilds, cause: errors.New("discord api not configured")}
	}
	userGuilds, err := h.discord.ListUserGuilds(ctx, userID)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return &discordGuildRestriction{eventTitle: title, eventNamed: named, guilds: guilds, cause: err}
	}
	if discord.MemberOfAny(userGuilds, required) {
		return nil
	}
	return &discordGuildRestriction{eventTitle: title, eventNamed: named, guilds: guilds}
}

// writeDiscordGuildRestriction writes the nested 403 envelope used by EventPage.
func (h *Handler) writeDiscordGuildRestriction(c *gin.Context, userID, groupID uuid.UUID, restricted *discordGuildRestriction) {
	if restricted.cause != nil {
		slog.ErrorContext(c.Request.Context(), "discord guild access check failed", "error", restricted.cause, "user_id", userID, "event_group_id", groupID)
	} else {
		slog.WarnContext(c.Request.Context(), "discord guild access denied", "user_id", userID, "event_group_id", groupID)
	}
	guilds := restricted.guilds
	if guilds == nil {
		guilds = []model.DiscordGuild{}
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"status":  "error",
		"message": "This event is locked to selected Discord servers.",
		"details": gin.H{
			"code":           "discord_guild_restricted",
			"event_title":    restricted.eventTitle,
			"event_named":    restricted.eventNamed,
			"discord_guilds": guilds,
		},
	})
}

// requireEventDiscordAllowed runs the lock check and writes 403/404/500 as needed. False means aborted.
func (h *Handler) requireEventDiscordAllowed(c *gin.Context, groupID, userID uuid.UUID) bool {
	err := h.isEventDiscordAllowed(c.Request.Context(), groupID, userID)
	if err == nil {
		return true
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		c.Abort()
		return false
	}
	var restricted *discordGuildRestriction
	if errors.As(err, &restricted) {
		h.writeDiscordGuildRestriction(c, userID, groupID, restricted)
		return false
	}
	if errors.Is(err, store.ErrEventGroupNotFound) || errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(c.Request.Context(), "event group not found for discord access check", "user_id", userID, "group_id", groupID, "error", err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event group not found"})
		return false
	}
	slog.ErrorContext(c.Request.Context(), "failed discord access check", "error", err, "user_id", userID, "event_group_id", groupID)
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to verify event access"})
	return false
}

// requireEventDiscordAllowedForEvent resolves the parent group, then applies the Discord lock.
func (h *Handler) requireEventDiscordAllowedForEvent(c *gin.Context, eventID, userID uuid.UUID) bool {
	groupID, err := h.store.EventGroupIDByEventID(c.Request.Context(), eventID)
	if err != nil {
		if errors.Is(err, store.ErrEventNotFound) || errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(c.Request.Context(), "event not found for discord access check", "user_id", userID, "event_id", eventID, "error", err)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Event not found"})
			return false
		}
		slog.ErrorContext(c.Request.Context(), "failed to resolve event group for discord access check", "error", err, "user_id", userID, "event_id", eventID)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to verify event access"})
		return false
	}
	return h.requireEventDiscordAllowed(c, groupID, userID)
}

// requireEventDiscordAllowedForLobby resolves the parent group from a lobby, then applies the Discord lock.
func (h *Handler) requireEventDiscordAllowedForLobby(c *gin.Context, lobbyID, userID uuid.UUID) bool {
	groupID, err := h.store.EventGroupIDByLobbyID(c.Request.Context(), lobbyID)
	if err != nil {
		if errors.Is(err, store.ErrLobbyNotFound) || errors.Is(err, pgx.ErrNoRows) {
			slog.WarnContext(c.Request.Context(), "lobby not found for discord access check", "user_id", userID, "lobby_id", lobbyID, "error", err)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"status": "error", "message": "Lobby not found"})
			return false
		}
		slog.ErrorContext(c.Request.Context(), "failed to resolve lobby group for discord access check", "error", err, "user_id", userID, "lobby_id", lobbyID)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to verify event access"})
		return false
	}
	return h.requireEventDiscordAllowed(c, groupID, userID)
}

// snapshotDiscordGuilds maps host-selected Discord snowflakes to name snapshots. Empty ids skip Discord.
func (h *Handler) snapshotDiscordGuilds(ctx context.Context, userID uuid.UUID, ids []string) ([]model.DiscordGuild, error) {
	cleaned := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		cleaned = append(cleaned, id)
	}
	if len(cleaned) == 0 {
		return []model.DiscordGuild{}, nil
	}
	if h.discord == nil {
		return nil, errors.New("discord api not configured")
	}
	userGuilds, err := h.discord.ListUserGuilds(ctx, userID)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]model.DiscordGuild, len(userGuilds))
	for _, g := range userGuilds {
		byID[g.ID] = g
	}
	out := make([]model.DiscordGuild, 0, len(cleaned))
	for _, id := range cleaned {
		g, ok := byID[id]
		if !ok {
			return nil, errDiscordGuildNotMember
		}
		out = append(out, model.DiscordGuild{ID: g.ID, Name: g.Name})
	}
	return out, nil
}

// writeDiscordGuildsLoadError maps ListUserGuilds failures for host picker / save paths.
func (h *Handler) writeDiscordGuildsLoadError(c *gin.Context, userID uuid.UUID, err error) {
	if errors.Is(err, discord.ErrMissingGrant) || errors.Is(err, discord.ErrInvalidGrant) {
		slog.ErrorContext(c.Request.Context(), "discord grant missing while loading guilds", "error", err, "user_id", userID)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Discord authorization expired. Please sign in again.",
		})
		return
	}
	slog.ErrorContext(c.Request.Context(), "failed to load discord guilds", "error", err, "user_id", userID)
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status":  "error",
		"message": "Could not load Discord servers",
	})
}
