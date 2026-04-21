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

func TestGetSystemGames_ReturnsGames(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		games, err := s.GetSystemGames(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, games)
	})
}

// ============================================================
// GetUserGames
// ============================================================

func TestGetUserGames_NilOwner_ReturnsOnlySystemGames(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		ctx := context.Background()
		games, err := s.GetUserGames(ctx, nil)
		require.NoError(t, err)
		require.NotEmpty(t, games, "migrations should seed system games")
		for _, g := range games {
			assert.Nil(t, g.OwnerID, "owner nil => only rows with owner_id IS NULL")
		}
	})
}

func TestGetUserGames_WithOwner_IncludesOwnedGame(t *testing.T) {
	pool := test_util.GetTestPool(t)
	defer pool.Close()
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	q := db.New(tx)
	s := store.NewPostgresStoreFromTx(tx)

	suffix := uuid.NewString()
	user := seedUser(t, q, "discord-gug-"+suffix, "gug-"+suffix)

	customID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO games (id, name, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())`,
		customID, "Custom Owned Game", user.ID,
	)
	require.NoError(t, err)

	games, err := s.GetUserGames(ctx, &user.ID)
	require.NoError(t, err)
	require.NotEmpty(t, games)

	var foundOwned bool
	systemCount := 0
	for _, g := range games {
		if g.ID == customID {
			foundOwned = true
			assert.Equal(t, "Custom Owned Game", g.Name)
			require.NotNil(t, g.OwnerID)
			assert.Equal(t, user.ID, *g.OwnerID)
		}
		if g.OwnerID == nil {
			systemCount++
		}
	}
	assert.True(t, foundOwned, "owned game should appear in GetUserGames")
	assert.Greater(t, systemCount, 0, "system games should still be included")
}

func TestGetUserGames_QueryError_nilOwnerInMessage(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := s.GetUserGames(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "owner (nil)")
	})
}

func TestGetUserGames_QueryError_ownerIDInMessage(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		owner := uuid.New()
		_, err := s.GetUserGames(ctx, &owner)
		require.Error(t, err)
		assert.Contains(t, err.Error(), owner.String())
	})
}

// ============================================================
// GetGameModeByID
// ============================================================

func TestGetGameModeByID_Found(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		ctx := context.Background()
		system, err := s.GetSystemGames(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, system)

		modes, err := s.GetGameModes(ctx, system[0].ID)
		require.NoError(t, err)
		require.NotEmpty(t, modes)

		want := modes[0]
		got, err := s.GetGameModeByID(ctx, want.ID)
		require.NoError(t, err)
		assert.Equal(t, want.ID, got.ID)
		assert.Equal(t, want.GameID, got.GameID)
		assert.Equal(t, want.Name, got.Name)
		assert.Equal(t, want.TeamSize, got.TeamSize)
		assert.Equal(t, want.Duration, got.Duration)
		assert.Equal(t, want.OwnerID, got.OwnerID)
		assert.Equal(t, want.CreatedAt, got.CreatedAt)
		assert.Equal(t, want.UpdatedAt, got.UpdatedAt)
	})
}

func TestGetGameModeByID_NotFound(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		_, err := s.GetGameModeByID(context.Background(), uuid.New())
		assert.Error(t, err)
	})
}
