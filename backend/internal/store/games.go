package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) GetSystemGames(ctx context.Context) ([]model.Game, error) {
	dbGame, err := s.q.GetSystemGames(ctx)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return []model.Game{}, fmt.Errorf("looking up system games: %w", err)
	}

	return model.MapDbGameToGame(dbGame), nil
}
