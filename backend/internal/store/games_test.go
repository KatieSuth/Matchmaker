package store_test

import (
	"context"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSystemGames_ReturnsGames(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		games, err := s.GetSystemGames(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, games)
	})
}
