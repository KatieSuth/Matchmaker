package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/lobbyjoin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrLobbyNotFound is returned when a lobby id does not exist.
var ErrLobbyNotFound = errors.New("lobby not found")

// UpdateLobbyJoinCode sets or clears a lobby's join_code. Actor must be the event-group
// owner or this lobby's host. joinCode may be a full URL, path suffix, plain code, or
// empty/nil to clear; values are normalized before storage.
func (s *PostgresStore) UpdateLobbyJoinCode(
	ctx context.Context,
	lobbyID, actorID uuid.UUID,
	rawJoinCode *string,
) error {
	auth, err := s.q.GetLobbyAuthContext(ctx, lobbyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLobbyNotFound
		}
		return fmt.Errorf("get lobby auth context: %w", err)
	}

	isOwner := auth.OwnerID == actorID
	isLobbyHost := auth.Host != nil && *auth.Host == actorID
	if !isOwner && !isLobbyHost {
		return ErrForbidden
	}

	raw := ""
	if rawJoinCode != nil {
		raw = *rawJoinCode
	}
	normalized, err := lobbyjoin.Normalize(raw, auth.JoinLinkBase)
	if err != nil {
		return err
	}

	if err := s.q.UpdateLobbyJoinCode(ctx, db.UpdateLobbyJoinCodeParams{
		ID:       lobbyID,
		JoinCode: normalized,
	}); err != nil {
		return fmt.Errorf("update lobby join code: %w", err)
	}
	return nil
}
