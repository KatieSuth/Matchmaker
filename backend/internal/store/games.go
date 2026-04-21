package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) GetSystemGames(ctx context.Context) ([]model.Game, error) {
	dbGame, err := s.q.GetSystemGames(ctx)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return []model.Game{}, fmt.Errorf("looking up system games: %w", err)
	}

	return model.MapDbGamesToGames(dbGame), nil
}

func (s *PostgresStore) GetUserGames(ctx context.Context, ownerID *uuid.UUID) ([]model.Game, error) {
	dbGames, err := s.q.GetUserGames(ctx, ownerID)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		owner := "nil"
		if ownerID != nil {
			owner = ownerID.String()
		}
		return []model.Game{}, fmt.Errorf("looking up user games for owner (%s): %w", owner, err)
	}

	return model.MapDbGamesToGames(dbGames), nil
}

func (s *PostgresStore) GetGameModes(ctx context.Context, gameID uuid.UUID) ([]model.GameMode, error) {
	dbGameModes, err := s.q.GetGameModes(ctx, gameID)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return []model.GameMode{}, fmt.Errorf("looking up game modes for game (%s): %w", gameID.String(), err)
	}

	return model.MapDbGameModesToGameModes(dbGameModes), nil
}

func (s *PostgresStore) GetGameModeByID(ctx context.Context, gameModeID uuid.UUID) (model.GameMode, error) {
	dbGameMode, err := s.q.GetGameModeById(ctx, gameModeID)
	if err != nil {
		return model.GameMode{}, fmt.Errorf("looking up game mode (%s): %w", gameModeID.String(), err)
	}

	return model.MapDbGameModeToGameMode(dbGameMode), nil
}
