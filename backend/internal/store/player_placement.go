package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func isSubPlacement(p swapPlacement) bool {
	return p.placed && p.teamNumber == nil
}

func isUnplacedPlacement(p swapPlacement) bool {
	return !p.placed
}

func isSubUnplacedSwap(a, b swapPlacement) bool {
	return (isSubPlacement(a) && isUnplacedPlacement(b)) || (isSubPlacement(b) && isUnplacedPlacement(a))
}

// MoveSubToUnplacedForEvent removes a substitute from their lobby sub pool.
func (s *PostgresStore) MoveSubToUnplacedForEvent(
	ctx context.Context,
	eventID, ownerID, userID uuid.UUID,
	settings matchmaking.Settings,
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
		return err
	}

	placements, err := s.q.GetPlayerPlacementsForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("get player placements: %w", err)
	}
	placementByUser := make(map[uuid.UUID]swapPlacement, len(placements))
	for _, row := range placements {
		placementByUser[row.UserID] = swapPlacement{
			placed:     true,
			lobbyID:    row.LobbyID,
			teamNumber: row.TeamNumber,
		}
	}

	place, ok := placementByUser[userID]
	if !ok || !isSubPlacement(place) {
		return &SwapValidationError{Message: "Player is not a substitute"}
	}

	lobbyRows, err := s.q.GetLobbiesForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("get lobbies for event: %w", err)
	}
	newPlace := swapPlacement{placed: false}
	if err := validateSubMinimumAfterSwap(placementByUser, userID, uuid.Nil, newPlace, swapPlacement{placed: false}, lobbyRows, int(meta.SubMin)); err != nil {
		return err
	}

	return s.WithTx(ctx, func(tx Store) error {
		txStore, ok := tx.(*PostgresStore)
		if !ok {
			return fmt.Errorf("unexpected tx store type %T", tx)
		}
		if err := txStore.q.DeletePlayer(ctx, db.DeletePlayerParams{
			LobbyID: place.lobbyID,
			UserID:  userID,
		}); err != nil {
			return fmt.Errorf("delete substitute player: %w", err)
		}
		if err := txStore.recomputeLobbyAfterSwap(ctx, place.lobbyID, settings, meta.GameModeID); err != nil {
			return err
		}
		return nil
	})
}

// MoveUnplacedToSubsForEvent adds an unplaced substitute-eligible player to a lobby sub pool.
func (s *PostgresStore) MoveUnplacedToSubsForEvent(
	ctx context.Context,
	eventID, ownerID, userID, lobbyID uuid.UUID,
	settings matchmaking.Settings,
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

	reg, err := s.loadSwapRegistration(ctx, eventID, userID)
	if err != nil {
		return err
	}
	if !reg.canSubstitute {
		return &SwapValidationError{Message: "Only substitute-eligible players can be moved to subs"}
	}

	placements, err := s.q.GetPlayerPlacementsForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("get player placements: %w", err)
	}
	for _, row := range placements {
		if row.UserID == userID {
			return &SwapValidationError{Message: "Player is already placed"}
		}
	}

	lobbyRows, err := s.q.GetLobbiesForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("get lobbies for event: %w", err)
	}
	var lobbyFound bool
	for _, lobby := range lobbyRows {
		if lobby.ID == lobbyID {
			lobbyFound = true
			break
		}
	}
	if !lobbyFound {
		return &SwapValidationError{Message: "Lobby not found for this game"}
	}

	return s.WithTx(ctx, func(tx Store) error {
		txStore, ok := tx.(*PostgresStore)
		if !ok {
			return fmt.Errorf("unexpected tx store type %T", tx)
		}
		if err := txStore.q.CreatePlayer(ctx, db.CreatePlayerParams{
			LobbyID:    lobbyID,
			UserID:     userID,
			TeamNumber: nil,
		}); err != nil {
			return fmt.Errorf("insert substitute player: %w", err)
		}
		if err := txStore.recomputeLobbyAfterSwap(ctx, lobbyID, settings, meta.GameModeID); err != nil {
			return err
		}
		return nil
	})
}
