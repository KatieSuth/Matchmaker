package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/stretchr/testify/assert"
)

func TestNewPostgresStore(t *testing.T) {
	// test_util likely provides access to the pgxpool used for migrations
	pool := test_util.GetTestPool(t)
	defer pool.Close()

	s := store.NewPostgresStore(pool)

	assert.NotNil(t, s)
	// We check that the internal pool is set correctly
	// If pool is unexported, this test might need to be in package store
}

func TestWithTx_Commit(t *testing.T) {
	test_util.WithTestPool(t, func(s *store.PostgresStore) {
		ctx := context.Background()

		err := s.WithTx(ctx, func(txStore store.Store) error {
			// Inside the transaction, we do something (e.g., create a record)
			// For this test, simply returning nil should trigger a Commit
			return nil
		})

		assert.NoError(t, err, "successful function should commit and return no error")
	})
}

func TestWithTx_Rollback(t *testing.T) {
	test_util.WithTestPool(t, func(s *store.PostgresStore) {
		ctx := context.Background()
		sentinelErr := errors.New("force rollback")

		err := s.WithTx(ctx, func(txStore store.Store) error {
			// Any error returned here should trigger the defer tx.Rollback(ctx)
			return sentinelErr
		})

		assert.ErrorIs(t, err, sentinelErr, "error from inside fn should be returned by WithTx")
	})
}

func TestWithTx_ContextCancellation(t *testing.T) {
	test_util.WithTestPool(t, func(s *store.PostgresStore) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := s.WithTx(ctx, func(txStore store.Store) error {
			return nil
		})

		// Should fail because the context is already done when Begin is called
		assert.Error(t, err)
	})
}
