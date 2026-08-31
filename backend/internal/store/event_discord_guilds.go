package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// replaceEventGroupDiscordGuilds deletes every lock row for the group, then inserts guilds.
// An empty slice leaves the group unrestricted. Must run inside the same transaction as create/update.
func (s *PostgresStore) replaceEventGroupDiscordGuilds(ctx context.Context, groupID uuid.UUID, guilds []model.DiscordGuild) error {
	if err := s.q.DeleteEventGroupDiscordGuildsByGroupID(ctx, groupID); err != nil {
		return fmt.Errorf("delete event group discord guilds: %w", err)
	}
	for _, guild := range guilds {
		_, err := s.q.InsertEventGroupDiscordGuild(ctx, db.InsertEventGroupDiscordGuildParams{
			EventGroupID: groupID,
			GuildID:      guild.ID,
			GuildName:    guild.Name,
		})
		if err != nil {
			return fmt.Errorf("insert event group discord guild %s: %w", guild.ID, err)
		}
	}
	return nil
}

// ListEventGroupDiscordGuilds returns the host-selected Discord servers for a group (empty = unrestricted).
func (s *PostgresStore) ListEventGroupDiscordGuilds(ctx context.Context, groupID uuid.UUID) ([]model.DiscordGuild, error) {
	rows, err := s.q.ListEventGroupDiscordGuildsByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("list event group discord guilds: %w", err)
	}
	return model.MapDbEventGroupDiscordGuildsToDiscordGuilds(rows), nil
}

// GetEventGroupAccessMeta returns the group owner, a display title, and whether the title is a host-assigned name.
func (s *PostgresStore) GetEventGroupAccessMeta(ctx context.Context, groupID uuid.UUID) (uuid.UUID, string, bool, error) {
	row, err := s.q.GetEventGroupDetailById(ctx, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, "", false, ErrEventGroupNotFound
		}
		return uuid.Nil, "", false, fmt.Errorf("get event group access meta: %w", err)
	}
	if row.Name != nil && *row.Name != "" {
		return row.OwnerID, *row.Name, true, nil
	}
	return row.OwnerID, row.GameName, false, nil
}

// EventGroupIDByEventID resolves the parent event group for a scheduled game.
func (s *PostgresStore) EventGroupIDByEventID(ctx context.Context, eventID uuid.UUID) (uuid.UUID, error) {
	meta, err := s.q.GetEventGroupMetaByEventId(ctx, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrEventNotFound
		}
		return uuid.Nil, fmt.Errorf("event group id by event: %w", err)
	}
	return meta.GroupID, nil
}

// EventGroupIDByLobbyID resolves the parent event group for a lobby.
func (s *PostgresStore) EventGroupIDByLobbyID(ctx context.Context, lobbyID uuid.UUID) (uuid.UUID, error) {
	auth, err := s.q.GetLobbyAuthContext(ctx, lobbyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrLobbyNotFound
		}
		return uuid.Nil, fmt.Errorf("event group id by lobby: %w", err)
	}
	return auth.GroupID, nil
}
