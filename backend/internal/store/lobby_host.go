package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrInvalidLobbyHostChange is the sentinel for lobby host validation failures surfaced to handlers.
var ErrInvalidLobbyHostChange = errors.New("invalid lobby host change")

// LobbyHostValidationError carries a client-facing message for lobby host validation failures.
type LobbyHostValidationError struct {
	Message string
}

func (e *LobbyHostValidationError) Error() string {
	return e.Message
}

func (e *LobbyHostValidationError) Unwrap() error {
	return ErrInvalidLobbyHostChange
}

// SetLobbyHostForEvent assigns a team player as the lobby host for their lobby.
// Host authorization and validation run here; can_lobby_host is not enforced server-side
// so the event owner may override a player's preference after confirming in the UI.
func (s *PostgresStore) SetLobbyHostForEvent(
	ctx context.Context,
	eventID, ownerID, userID uuid.UUID,
) error {
	meta, err := s.q.GetEventGroupMetaByEventId(ctx, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEventNotFound
		}
		return fmt.Errorf("get event group meta: %w", err)
	}
	if meta.OwnerID != ownerID {
		return ErrForbidden
	}

	lobbyCount, err := s.q.CountLobbiesForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("count lobbies for event: %w", err)
	}
	if lobbyCount == 0 {
		return ErrTeamsNotCreated
	}

	if _, err := s.loadSwapRegistration(ctx, eventID, userID); err != nil {
		var swapErr *SwapValidationError
		if errors.As(err, &swapErr) {
			return &LobbyHostValidationError{Message: swapErr.Message}
		}
		return err
	}

	placements, err := s.q.GetPlayerPlacementsForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("get player placements: %w", err)
	}

	var placement *swapPlacement
	for _, row := range placements {
		if row.UserID != userID {
			continue
		}
		placement = &swapPlacement{
			placed:     true,
			lobbyID:    row.LobbyID,
			teamNumber: row.TeamNumber,
		}
		break
	}
	if placement == nil {
		return &LobbyHostValidationError{Message: "Player is not assigned to a team"}
	}
	if placement.teamNumber == nil {
		return &LobbyHostValidationError{Message: "Only team players can be made lobby host"}
	}

	lobbyRows, err := s.q.GetLobbiesForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("get lobbies for event: %w", err)
	}

	var lobbyHost *uuid.UUID
	for _, lobby := range lobbyRows {
		if lobby.ID != placement.lobbyID {
			continue
		}
		lobbyHost = lobby.Host
		break
	}
	if lobbyHost == nil {
		return &LobbyHostValidationError{Message: "Lobby not found for this player"}
	}
	if *lobbyHost == userID {
		return &LobbyHostValidationError{Message: "Player is already the lobby host"}
	}

	hostID := userID
	if err := s.q.UpdateLobbyHost(ctx, db.UpdateLobbyHostParams{
		ID:   placement.lobbyID,
		Host: &hostID,
	}); err != nil {
		return fmt.Errorf("update lobby host: %w", err)
	}

	return nil
}
