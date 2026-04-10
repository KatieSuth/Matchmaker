package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) GetGameRanks(ctx context.Context, gameID *uuid.UUID) ([]model.GameRank, error) {
	dbGameRanks, err := s.q.GetRanksForGame(ctx, gameID)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return []model.GameRank{}, fmt.Errorf("looking up ranks for game (%s): %w", gameID.String(), err)
	}

	return model.MapDbGameRanksToGameRanks(dbGameRanks), nil
}
