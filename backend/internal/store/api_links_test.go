package store_test

import (
	"context"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/apilink"
	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertApiLink_InsertAndGet(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-al1", "aluser1")

		link, err := s.UpsertApiLink(context.Background(), seeded.ID, apilink.ProviderDiscord, "cipher-1", "nonce-1", "1")
		require.NoError(t, err)
		assert.Equal(t, seeded.ID, link.UserID)
		assert.Equal(t, apilink.ProviderDiscord, link.Name)
		assert.Equal(t, "cipher-1", link.RefreshToken)
		assert.Equal(t, "nonce-1", link.RefreshTokenIv)
		assert.Equal(t, "1", link.KeyID)

		got, err := s.GetApiLinkByUserAndName(context.Background(), seeded.ID, apilink.ProviderDiscord)
		require.NoError(t, err)
		assert.Equal(t, link.ID, got.ID)
		assert.Equal(t, "cipher-1", got.RefreshToken)
	})
}

func TestUpsertApiLink_OverwritesSameName(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-al2", "aluser2")

		first, err := s.UpsertApiLink(context.Background(), seeded.ID, apilink.ProviderDiscord, "cipher-old", "nonce-old", "1")
		require.NoError(t, err)

		second, err := s.UpsertApiLink(context.Background(), seeded.ID, apilink.ProviderDiscord, "cipher-new", "nonce-new", "2")
		require.NoError(t, err)
		assert.Equal(t, first.ID, second.ID)
		assert.Equal(t, "cipher-new", second.RefreshToken)
		assert.Equal(t, "nonce-new", second.RefreshTokenIv)
		assert.Equal(t, "2", second.KeyID)
	})
}

func TestUpsertApiLink_TwoProvidersDoNotClobber(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-al3", "aluser3")

		_, err := s.UpsertApiLink(context.Background(), seeded.ID, apilink.ProviderDiscord, "discord-cipher", "discord-nonce", "1")
		require.NoError(t, err)
		_, err = s.UpsertApiLink(context.Background(), seeded.ID, "riot", "riot-cipher", "riot-nonce", "1")
		require.NoError(t, err)

		discord, err := s.GetApiLinkByUserAndName(context.Background(), seeded.ID, apilink.ProviderDiscord)
		require.NoError(t, err)
		assert.Equal(t, "discord-cipher", discord.RefreshToken)

		riot, err := s.GetApiLinkByUserAndName(context.Background(), seeded.ID, "riot")
		require.NoError(t, err)
		assert.Equal(t, "riot-cipher", riot.RefreshToken)
	})
}

func TestGetApiLinkByUserAndName_NotFound(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-al4", "aluser4")
		_, err := s.GetApiLinkByUserAndName(context.Background(), seeded.ID, apilink.ProviderDiscord)
		assert.Error(t, err)
	})
}

func TestApiLink_CascadesOnUserDelete(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	user := createTestUser(t, ctx, s)

	_, err := s.UpsertApiLink(ctx, user.ID, apilink.ProviderDiscord, "cipher", "nonce", "1")
	require.NoError(t, err)

	_, err = tx.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	require.NoError(t, err)

	_, err = s.GetApiLinkByUserAndName(ctx, user.ID, apilink.ProviderDiscord)
	assert.Error(t, err)
}

func TestUpsertApiLink_QueryError(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-al-err", "alerr")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := s.UpsertApiLink(ctx, seeded.ID, apilink.ProviderDiscord, "c", "n", "1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "upserting api link")
	})
}

func TestGetApiLinkByUserAndNameForUpdate_Success(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-al-fu", "alfu")
		_, err := s.UpsertApiLink(context.Background(), seeded.ID, apilink.ProviderDiscord, "cipher-fu", "nonce-fu", "1")
		require.NoError(t, err)

		got, err := s.GetApiLinkByUserAndNameForUpdate(context.Background(), seeded.ID, apilink.ProviderDiscord)
		require.NoError(t, err)
		assert.Equal(t, "cipher-fu", got.RefreshToken)
	})
}

func TestGetApiLinkByUserAndNameForUpdate_QueryError(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-al-fu-err", "alfuerr")
		_, err := s.UpsertApiLink(context.Background(), seeded.ID, apilink.ProviderDiscord, "cipher-fu", "nonce-fu", "1")
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = s.GetApiLinkByUserAndNameForUpdate(ctx, seeded.ID, apilink.ProviderDiscord)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "locating api link for update")
	})
}

func TestDeleteApiLinkByUserAndName_Success(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-al-del", "aldel")
		_, err := s.UpsertApiLink(context.Background(), seeded.ID, apilink.ProviderDiscord, "cipher-del", "nonce-del", "1")
		require.NoError(t, err)

		err = s.DeleteApiLinkByUserAndName(context.Background(), seeded.ID, apilink.ProviderDiscord)
		require.NoError(t, err)

		_, err = s.GetApiLinkByUserAndName(context.Background(), seeded.ID, apilink.ProviderDiscord)
		assert.Error(t, err)
	})
}

func TestDeleteApiLinkByUserAndName_QueryError(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-al-del-err", "aldelerr")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := s.DeleteApiLinkByUserAndName(ctx, seeded.ID, apilink.ProviderDiscord)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleting api link")
	})
}
