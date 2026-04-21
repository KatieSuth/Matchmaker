package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Raw SQL below verifies rows because there is no store API to list events by group yet.
func createEventTestStoreTx(t *testing.T) (*store.PostgresStore, db.DBTX) {
	t.Helper()
	pool := test_util.GetTestPool(t)
	t.Cleanup(func() { pool.Close() })

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	return store.NewPostgresStoreFromTx(tx), tx
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
	s, tx := createEventTestStoreTx(t)
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
	err = tx.QueryRow(ctx, `
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

	rows, err := tx.Query(ctx, `
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
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()

	user := createTestUser(t, ctx, s)

	_, err := s.CreateEventGroupWithEvents(ctx, user.ID, uuid.New(), 1, false, "NA", time.Now().UTC(), 1)
	require.Error(t, err)
}

func TestCreateEventGroupWithEvents_SingleGame(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
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
	err = tx.QueryRow(ctx, `SELECT count(*) FROM events WHERE group_id = $1`, groupID).Scan(&n)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

func firstModeForGame(t *testing.T, ctx context.Context, s *store.PostgresStore, gameID uuid.UUID) model.GameMode {
	t.Helper()
	modes, err := s.GetGameModes(ctx, gameID)
	require.NoError(t, err)
	require.NotEmpty(t, modes)
	return modes[0]
}

func insertEventFixture(t *testing.T, ctx context.Context, tx db.DBTX, ownerID, gameModeID uuid.UUID, start time.Time) (uuid.UUID, uuid.UUID) {
	t.Helper()

	groupID := uuid.New()
	eventID := uuid.New()
	_, err := tx.Exec(ctx, `
		INSERT INTO event_groups (id, owner_id, game_mode_id, sub_min, registration_open, region, created_at, updated_at)
		VALUES ($1, $2, $3, 0, true, 'NA', NOW(), NOW())
	`, groupID, ownerID, gameModeID)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `
		INSERT INTO events (id, group_id, start_time, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, eventID, groupID, start.UTC())
	require.NoError(t, err)

	return groupID, eventID
}

func insertEventInGroupFixture(t *testing.T, ctx context.Context, tx db.DBTX, groupID uuid.UUID, start time.Time) uuid.UUID {
	t.Helper()

	eventID := uuid.New()
	_, err := tx.Exec(ctx, `
		INSERT INTO events (id, group_id, start_time, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, eventID, groupID, start.UTC())
	require.NoError(t, err)

	return eventID
}

func registerUserForEvent(t *testing.T, ctx context.Context, tx db.DBTX, eventID, userID uuid.UUID) {
	t.Helper()
	_, err := tx.Exec(ctx, `
		INSERT INTO registrations (event_id, user_id, can_substitute, can_lobby_host, created_at, updated_at)
		VALUES ($1, $2, false, false, NOW(), NOW())
	`, eventID, userID)
	require.NoError(t, err)
}

func TestGetEventsForUser_FiltersAndTimezoneBoundaries(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("America/New_York timezone unavailable")
	}

	host := createTestUser(t, ctx, s)
	participant := createTestUser(t, ctx, s)

	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(games), 2)

	modeA := firstModeForGame(t, ctx, s, games[0].ID)
	modeB := firstModeForGame(t, ctx, s, games[1].ID)

	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	pastStart := todayStart.Add(-2 * time.Hour)
	upcomingStart := todayStart.Add(2 * time.Hour)
	otherGameStart := todayStart.Add(3 * time.Hour)

	groupA, pastEvent := insertEventFixture(t, ctx, tx, host.ID, modeA.ID, pastStart)
	upcomingEvent := insertEventInGroupFixture(t, ctx, tx, groupA, upcomingStart)
	_, _ = insertEventFixture(t, ctx, tx, host.ID, modeB.ID, otherGameStart)

	registerUserForEvent(t, ctx, tx, pastEvent, participant.ID)
	registerUserForEvent(t, ctx, tx, upcomingEvent, participant.ID)

	upcomingHosted, hasMore, nextCursor, err := s.GetEventsForUser(
		ctx, host.ID, true, false, nil, nil, games[0].ID.String(), "", "America/New_York",
	)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Empty(t, nextCursor)
	require.Len(t, upcomingHosted, 1)
	assert.Equal(t, groupA, upcomingHosted[0].ID)

	pastHosted, hasMore, nextCursor, err := s.GetEventsForUser(
		ctx, host.ID, true, true, nil, nil, games[0].ID.String(), "", "America/New_York",
	)
	require.NoError(t, err)
	require.False(t, hasMore)
	require.Empty(t, nextCursor)
	require.Len(t, pastHosted, 1)
	assert.Equal(t, groupA, pastHosted[0].ID)

	upcomingRegistered, _, _, err := s.GetEventsForUser(
		ctx, participant.ID, false, false, nil, nil, games[0].ID.String(), "", "America/New_York",
	)
	require.NoError(t, err)
	require.Len(t, upcomingRegistered, 1)
	assert.Equal(t, groupA, upcomingRegistered[0].ID)

	from := todayStart.AddDate(0, 0, -1)
	to := from
	rangeResult, _, _, err := s.GetEventsForUser(
		ctx, host.ID, true, false, &from, &to, games[0].ID.String(), "", "America/New_York",
	)
	require.NoError(t, err)
	require.Len(t, rangeResult, 1)
	assert.Equal(t, groupA, rangeResult[0].ID)
}

func TestGetEventsForUser_CursorPagination(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("America/New_York timezone unavailable")
	}

	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)
	mode := firstModeForGame(t, ctx, s, games[0].ID)

	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	for i := 0; i < 22; i++ {
		_, _ = insertEventFixture(t, ctx, tx, host.ID, mode.ID, todayStart.Add(time.Duration(i)*time.Minute))
	}

	page1, hasMore, nextCursor, err := s.GetEventsForUser(
		ctx, host.ID, true, false, nil, nil, games[0].ID.String(), "", "America/New_York",
	)
	require.NoError(t, err)
	require.True(t, hasMore)
	require.NotEmpty(t, nextCursor)
	require.Len(t, page1, 20)

	page2, hasMore2, nextCursor2, err := s.GetEventsForUser(
		ctx, host.ID, true, false, nil, nil, games[0].ID.String(), nextCursor, "America/New_York",
	)
	require.NoError(t, err)
	require.False(t, hasMore2)
	require.Empty(t, nextCursor2)
	require.Len(t, page2, 2)
	assert.NotEqual(t, page1[len(page1)-1].ID, page2[0].ID)
}
