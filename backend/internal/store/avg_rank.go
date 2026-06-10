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

// resolveAvgRankID looks up the game_ranks row for the floored average of current and peak orders.
func (s *PostgresStore) resolveAvgRankID(ctx context.Context, gameID uuid.UUID, currentOrder, peakOrder int32) (uuid.UUID, int32, error) {
	flooredOrder := int32(matchmaking.FlooredAverageRankOrder(int(currentOrder), int(peakOrder)))
	rank, err := s.q.GetRankByGameAndOrder(ctx, db.GetRankByGameAndOrderParams{
		GameID: &gameID,
		Order:  flooredOrder,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, 0, fmt.Errorf("no rank for game %s at order %d", gameID.String(), flooredOrder)
		}
		return uuid.Nil, 0, fmt.Errorf("lookup avg rank for game %s order %d: %w", gameID.String(), flooredOrder, err)
	}
	return rank.ID, flooredOrder, nil
}

// ensureAvgRanksForMatchmaking backfills missing avg_rank values and returns rows ready for planning.
func (s *PostgresStore) ensureAvgRanksForMatchmaking(ctx context.Context, rows []db.GetMatchmakingRegistrationsForEventRow) ([]db.GetMatchmakingRegistrationsForEventRow, error) {
	for i := range rows {
		if rows[i].AvgRank != nil && *rows[i].AvgRank != uuid.Nil && rows[i].AvgRankOrder != nil {
			continue
		}

		avgRankID, avgOrder, err := s.resolveAvgRankID(ctx, rows[i].GameID, rows[i].CurrentRankOrder, rows[i].PeakRankOrder)
		if err != nil {
			return nil, err
		}

		if err := s.q.UpdateUserGameAvgRank(ctx, db.UpdateUserGameAvgRankParams{
			AvgRank: &avgRankID,
			UserID:  rows[i].UserID,
			GameID:  rows[i].GameID,
		}); err != nil {
			return nil, fmt.Errorf("update avg rank for user %s game %s: %w", rows[i].UserID.String(), rows[i].GameID.String(), err)
		}

		rows[i].AvgRank = &avgRankID
		rows[i].AvgRankOrder = &avgOrder
	}
	return rows, nil
}
