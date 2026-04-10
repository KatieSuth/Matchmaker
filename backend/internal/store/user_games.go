package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) GetUserGamesForUser(ctx context.Context, userID uuid.UUID) ([]model.UserGame, error) {
	dbUserGames, err := s.q.GetGamesForUser(ctx, userID)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return []model.UserGame{}, fmt.Errorf("looking up user's games: %w", err)
	}

	return model.MapDbUserGamesToUserGames(dbUserGames), nil
}
