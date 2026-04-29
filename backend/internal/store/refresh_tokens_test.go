package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNewRefreshToken_Success(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-rt1", "rtuser1")
		expires := time.Now().Add(7 * 24 * time.Hour)

		rt, err := s.CreateNewRefreshToken(context.Background(), "token-hash-1", seeded.ID, expires)
		require.NoError(t, err)
		assert.Equal(t, "token-hash-1", rt.Token)
		assert.Equal(t, seeded.ID, rt.UserID)
		assert.WithinDuration(t, expires, rt.ExpiresAt, time.Second)
	})
}

func TestGetRefreshToken_Found(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-rt2", "rtuser2")
		expires := time.Now().Add(7 * 24 * time.Hour)
		_, err := s.CreateNewRefreshToken(context.Background(), "token-hash-2", seeded.ID, expires)
		require.NoError(t, err)

		rt, err := s.GetRefreshToken(context.Background(), "token-hash-2")
		require.NoError(t, err)
		assert.Equal(t, "token-hash-2", rt.Token)
		assert.Equal(t, seeded.ID, rt.UserID)
	})
}

func TestGetRefreshToken_NotFound(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		_, err := s.GetRefreshToken(context.Background(), "nonexistent-hash")
		assert.Error(t, err)
	})
}

func TestDeleteRefreshToken_Success(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-rt3", "rtuser3")
		expires := time.Now().Add(7 * 24 * time.Hour)
		_, err := s.CreateNewRefreshToken(context.Background(), "token-hash-3", seeded.ID, expires)
		require.NoError(t, err)

		err = s.DeleteRefreshToken(context.Background(), "token-hash-3")
		require.NoError(t, err)

		// Confirm it's gone.
		_, err = s.GetRefreshToken(context.Background(), "token-hash-3")
		assert.Error(t, err)
	})
}

func TestDeleteRefreshToken_NotFound(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		// We expect no error even if the hash doesn't exist
		err := s.DeleteRefreshToken(context.Background(), "nonexistent-hash")
		assert.NoError(t, err)
	})
}


func TestCreateNewRefreshToken_QueryError(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-rt-err1", "rt-err1")
		expires := time.Now().Add(24 * time.Hour)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := s.CreateNewRefreshToken(ctx, "token-hash-err", seeded.ID, expires)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "creating refresh token")
	})
}

func TestDeleteRefreshToken_QueryError(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := s.DeleteRefreshToken(ctx, "token-hash-err")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleting refresh token")
	})
}
