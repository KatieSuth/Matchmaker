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

func TestCreateOneTimeCode_Success(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-otc1", "otcuser1")

		err := s.CreateOneTimeCode(context.Background(), "otc-code-1", seeded.ID)
		require.NoError(t, err)
	})
}

func TestConsumeOneTimeCode_ValidCode(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-otc2", "otcuser2")
		err := s.CreateOneTimeCode(context.Background(), "otc-code-2", seeded.ID)
		require.NoError(t, err)

		userID, err := s.ConsumeOneTimeCode(context.Background(), "otc-code-2")
		require.NoError(t, err)
		assert.Equal(t, seeded.ID, userID)
	})
}

func TestConsumeOneTimeCode_InvalidCode(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		// ConsumeOneTimeCode returns uuid.Nil (not an error) for missing codes
		// per the pgx.ErrNoRows handling in the implementation.
		userID, err := s.ConsumeOneTimeCode(context.Background(), "nonexistent-code")
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, userID)
	})
}

func TestConsumeOneTimeCode_CannotConsumeCodeTwice(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-otc3", "otcuser3")
		err := s.CreateOneTimeCode(context.Background(), "otc-code-3", seeded.ID)
		require.NoError(t, err)

		// First consume succeeds.
		userID, err := s.ConsumeOneTimeCode(context.Background(), "otc-code-3")
		require.NoError(t, err)
		assert.Equal(t, seeded.ID, userID)

		// Second consume returns nil UUID — code was deleted.
		userID, err = s.ConsumeOneTimeCode(context.Background(), "otc-code-3")
		require.NoError(t, err)
		assert.Equal(t, uuid.Nil, userID)
	})
}
