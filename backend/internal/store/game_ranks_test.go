package store_test

import (
	"context"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGameRanks_UnknownGame(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		gameID := uuid.New()
		ranks, err := s.GetGameRanks(context.Background(), &gameID)
		require.NoError(t, err)
		assert.Empty(t, ranks)
	})
}

func TestGetGameRanks_KnownGame(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		games, err := s.GetSystemGames(context.Background())
		require.NoError(t, err)

		if len(games) == 0 {
			t.Skip("no system games seeded — skipping")
		}

		ranks, err := s.GetGameRanks(context.Background(), &games[0].ID)
		require.NoError(t, err)
		assert.NotNil(t, ranks)
	})
}
