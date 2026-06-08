package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedUser creates a user directly via sqlc for use as test fixtures.
func seedUser(t *testing.T, q *db.Queries, discordID, discordName string) db.User {
	t.Helper()
	user, err := q.CreateUser(context.Background(), db.CreateUserParams{
		DiscordID:   &discordID,
		DiscordName: &discordName,
	})
	require.NoError(t, err, "failed to seed user")
	return user
}

// ============================================================
// GetUserByDiscordID
// ============================================================

func TestGetUserByDiscordID_Found(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-123", "testuser")

		user, err := s.GetUserByDiscordID(context.Background(), "discord-123", true)
		require.NoError(t, err)
		assert.Equal(t, seeded.ID, user.ID)
	})
}

func TestGetUserByDiscordID_NotFound(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		_, err := s.GetUserByDiscordID(context.Background(), "nonexistent-id", true)
		assert.Error(t, err)
	})
}

func TestGetUserByDiscordID_QueryError(t *testing.T) {
	_, tx := createEventTestStoreTx(t)
	injectedErr := errors.New("injected database failure")
	faulty := &faultInjectTx{DBTX: tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
	fs := store.NewPostgresStoreFromDBTXForTest(faulty)

	_, err := fs.GetUserByDiscordID(context.Background(), "discord-any", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "looking up user")
}

// ============================================================
// GetUserByUserID
// ============================================================

func TestGetUserByUserID_Found(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-456", "anotheruser")

		user, err := s.GetUserByUserID(context.Background(), seeded.ID)
		require.NoError(t, err)
		assert.Equal(t, seeded.ID, user.ID)
	})
}

func TestGetUserByUserID_NotFound(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		_, err := s.GetUserByUserID(context.Background(), uuid.New())
		assert.Error(t, err)
	})
}

// ============================================================
// CreateNewUser
// ============================================================

func TestCreateNewUser_Success(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		discordUser := model.DiscordUser{
			ID:       "discord-new-123",
			Username: "newuser",
			Avatar:   "avatar-hash",
		}

		user, err := s.CreateNewUser(context.Background(), discordUser)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.Equal(t, &discordUser.ID, user.DiscordID)
		assert.Equal(t, &discordUser.Username, user.DiscordName)
		assert.Equal(t, &discordUser.Avatar, user.ImageUrl)
	})
}

func TestCreateNewUser_DuplicateDiscordID(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seedUser(t, q, "discord-dup", "firstuser")

		_, err := s.CreateNewUser(context.Background(), model.DiscordUser{
			ID:       "discord-dup",
			Username: "seconduser",
		})
		assert.Error(t, err)
	})
}

// ============================================================
// UpdateUserFromLogin
// ============================================================

func TestUpdateUserFromLogin_UpdatesNameAndAvatar(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-upd", "oldname")

		updated, err := s.UpdateUserFromLogin(context.Background(), seeded.ID, model.DiscordUser{
			ID:       "discord-upd",
			Username: "newname",
			Avatar:   "new-avatar",
		})
		require.NoError(t, err)
		assert.Equal(t, "newname", *updated.DiscordName)
		assert.Equal(t, "new-avatar", *updated.ImageUrl)
	})
}

func TestUpdateUserFromLogin_NotFound(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		_, err := s.UpdateUserFromLogin(context.Background(), uuid.New(), model.DiscordUser{
			ID:       "ghost",
			Username: "nobody",
			Avatar:   "none",
		})
		require.Error(t, err)
	})
}

// ============================================================
// UpdateUser
// ============================================================

func TestUpdateUser_UpdatesPronouns(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-upd2", "pronounuser")
		pronouns := "they/them"
		region := "EU"

		updated, err := s.UpdateUser(context.Background(), seeded.ID, &pronouns, true, &region)
		require.NoError(t, err)
		assert.Equal(t, &pronouns, updated.Pronouns)
		assert.True(t, updated.ShowPronouns)
		assert.Equal(t, &region, updated.Region)
	})
}

func TestUpdateUser_NotFound(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		pronouns := "she/her"
		region := "AMER"
		_, err := s.UpdateUser(context.Background(), uuid.New(), &pronouns, false, &region)
		assert.Error(t, err)
	})
}
