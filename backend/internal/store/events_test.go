package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CreateEventGroupWithEvents calls WithTx, which uses pool.Begin. test_util.WithTestTx wires
// store.NewPostgresStoreFromTx(tx) with pool == nil, so WithTx would panic. These tests use
// test_util.GetTestPool like TestNewPostgresStore / TestWithTx_Commit in postgres_test.go.
// Raw SQL below verifies rows because there is no store API to list events by group yet.

func createEventTestStore(t *testing.T) (*pgxpool.Pool, *store.PostgresStore) {
	t.Helper()
	pool := test_util.GetTestPool(t)
	t.Cleanup(func() { pool.Close() })
	return pool, store.NewPostgresStore(pool)
}

// createTestUser mirrors users_test seedUser intent: a unique user for integration fixtures.
// It uses CreateNewUser so we only need the exported store API (no *db.Queries in scope).
func createTestUser(t *testing.T, ctx context.Context, s *store.PostgresStore) model.User {
	t.Helper()
	suffix := uuid.NewString()
	user, err := s.CreateNewUser(ctx, model.DiscordUser{
		ID:       "discord-cev-" + suffix,
		Username: "cevtest-" + suffix,
		Avatar:   "avatar",
	})
	require.NoError(t, err, "failed to create test user")
	return user
}

func TestCreateEventGroupWithEvents_Success(t *testing.T) {
	pool, s := createEventTestStore(t)
	ctx := context.Background()

	user := createTestUser(t, ctx, s)

	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games, "migrations should seed at least one system game")

	modes, err := s.GetGameModes(ctx, games[0].ID)
	require.NoError(t, err)
	require.NotEmpty(t, modes)

	var mode model.GameMode
	for _, m := range modes {
		if m.Duration > 0 {
			mode = m
			break
		}
	}
	require.NotEqual(t, uuid.Nil, mode.ID, "need a game mode with positive duration for schedule test")

	start := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	gamesToRun := int32(3)
	subMin := int32(2)

	groupID, err := s.CreateEventGroupWithEvents(ctx, user.ID, mode.ID, subMin, true, "EMEA", start, gamesToRun)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, groupID)

	var region string
	var gotSubMin int32
	var regOpen bool
	var ownerID uuid.UUID
	var gameModeID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT region, sub_min, registration_open, owner_id, game_mode_id
		FROM event_groups WHERE id = $1`,
		groupID,
	).Scan(&region, &gotSubMin, &regOpen, &ownerID, &gameModeID)
	require.NoError(t, err)
	assert.Equal(t, "EMEA", region)
	assert.Equal(t, subMin, gotSubMin)
	assert.True(t, regOpen)
	assert.Equal(t, user.ID, ownerID)
	assert.Equal(t, mode.ID, gameModeID)

	rows, err := pool.Query(ctx, `
		SELECT start_time FROM events WHERE group_id = $1 ORDER BY start_time ASC`,
		groupID,
	)
	require.NoError(t, err)
	defer rows.Close()

	var starts []time.Time
	for rows.Next() {
		var st time.Time
		require.NoError(t, rows.Scan(&st))
		starts = append(starts, st.UTC())
	}
	require.NoError(t, rows.Err())
	require.Len(t, starts, int(gamesToRun))

	for i := range starts {
		want := start.Add(time.Duration(mode.Duration) * time.Minute * time.Duration(i))
		assert.Equal(t, want.Unix(), starts[i].Unix(), "event %d start_time", i)
	}
}

func TestCreateEventGroupWithEvents_UnknownGameMode(t *testing.T) {
	_, s := createEventTestStore(t)
	ctx := context.Background()

	user := createTestUser(t, ctx, s)

	_, err := s.CreateEventGroupWithEvents(ctx, user.ID, uuid.New(), 1, false, "NA", time.Now().UTC(), 1)
	require.Error(t, err)
}

func TestCreateEventGroupWithEvents_SingleGame(t *testing.T) {
	pool, s := createEventTestStore(t)
	ctx := context.Background()

	user := createTestUser(t, ctx, s)

	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)

	modes, err := s.GetGameModes(ctx, games[0].ID)
	require.NoError(t, err)
	require.NotEmpty(t, modes)
	mode := modes[0]

	start := time.Date(2026, 5, 1, 9, 30, 0, 0, time.UTC)
	groupID, err := s.CreateEventGroupWithEvents(ctx, user.ID, mode.ID, 1, true, "APAC", start, 1)
	require.NoError(t, err)

	var n int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM events WHERE group_id = $1`, groupID).Scan(&n)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}
