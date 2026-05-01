package store_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	groupID, err := s.CreateEventGroupWithEvents(ctx, user.ID, mode.ID, subMin, true, "EMEA", "ranked", start, gamesToRun)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, groupID)

	var region string
	var gotSubMin int32
	var regOpen bool
	var ownerID uuid.UUID
	var gameModeID uuid.UUID
	var gotSort string
	err = tx.QueryRow(ctx, `
		SELECT region, sub_min, registration_open, owner_id, game_mode_id, sort_logic
		FROM event_groups WHERE id = $1`,
		groupID,
	).Scan(&region, &gotSubMin, &regOpen, &ownerID, &gameModeID, &gotSort)
	require.NoError(t, err)
	assert.Equal(t, "ranked", gotSort)
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

	_, err := s.CreateEventGroupWithEvents(ctx, user.ID, uuid.New(), 1, false, "AMER", "balanced", time.Now().UTC(), 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrGameModeNotFound)
}

func TestCreateEventGroupWithEvents_InvalidSortLogic(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	user := createTestUser(t, ctx, s)

	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)
	modes, err := s.GetGameModes(ctx, games[0].ID)
	require.NoError(t, err)
	require.NotEmpty(t, modes)

	_, err = s.CreateEventGroupWithEvents(ctx, user.ID, modes[0].ID, 0, true, "AMER", "lobster", time.Now().UTC().Add(24*time.Hour), 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidSortLogic)
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
	groupID, err := s.CreateEventGroupWithEvents(ctx, user.ID, mode.ID, 1, true, "APAC", "balanced", start, 1)
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

func insertExtraGameMode(t *testing.T, ctx context.Context, tx db.DBTX, gameID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := tx.Exec(ctx, `
		INSERT INTO game_modes (id, game_id, "name", team_size, created_at, updated_at, duration)
		VALUES ($1, $2, 'Alternate mode', 5, NOW(), NOW(), 60)
	`, id, gameID)
	require.NoError(t, err)
	return id
}

func createCompleteUserGameForRegistration(t *testing.T, ctx context.Context, tx db.DBTX, s *store.PostgresStore, userID, gameID uuid.UUID) {
	t.Helper()

	var rankID uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM game_ranks WHERE game_id = $1 ORDER BY "order" ASC LIMIT 1`, gameID).Scan(&rankID)
	require.NoError(t, err)

	inGameName := "registered-player"
	_, err = s.UpsertGameForUser(ctx, userID, model.UserGame{
		GameID:      gameID,
		InGameName:  &inGameName,
		CurrentRank: &rankID,
		PeakRank:    &rankID,
		ShowRank:    true,
	})
	require.NoError(t, err)
}

func insertEventFixture(t *testing.T, ctx context.Context, tx db.DBTX, ownerID, gameModeID uuid.UUID, start time.Time) (uuid.UUID, uuid.UUID) {
	t.Helper()

	groupID := uuid.New()
	eventID := uuid.New()
	_, err := tx.Exec(ctx, `
		INSERT INTO event_groups (id, owner_id, game_mode_id, sub_min, registration_open, region, created_at, updated_at)
		VALUES ($1, $2, $3, 0, true, 'AMER', NOW(), NOW())
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
	registerUserForEventWithFlags(t, ctx, tx, eventID, userID, false, false)
}

func registerUserForEventWithFlags(t *testing.T, ctx context.Context, tx db.DBTX, eventID, userID uuid.UUID, canSubstitute, canLobbyHost bool) {
	t.Helper()
	_, err := tx.Exec(ctx, `
		INSERT INTO registrations (event_id, user_id, can_substitute, can_lobby_host, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, eventID, userID, canSubstitute, canLobbyHost)
	require.NoError(t, err)
}

func insertLobbyForEvent(t *testing.T, ctx context.Context, tx db.DBTX, eventID uuid.UUID, host *uuid.UUID) uuid.UUID {
	t.Helper()
	lobbyID := uuid.New()
	_, err := tx.Exec(ctx, `
		INSERT INTO lobbies (id, event_id, host, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, lobbyID, eventID, host)
	require.NoError(t, err)
	return lobbyID
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

func TestGetEventsForUser_InvalidInputs(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	user := createTestUser(t, ctx, s)

	_, _, _, err := s.GetEventsForUser(ctx, user.ID, true, false, nil, nil, "", "", "Not/A_Real_Timezone")
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidTimezone)

	_, _, _, err = s.GetEventsForUser(ctx, user.ID, true, false, nil, nil, "not-a-uuid", "", "UTC")
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidGameID)

	_, _, _, err = s.GetEventsForUser(ctx, user.ID, true, false, nil, nil, "", "not-base64", "UTC")
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidCursor)

	payload, marshalErr := json.Marshal(map[string]any{
		"event_date": time.Now().UTC(),
		"id":         uuid.Nil,
	})
	require.NoError(t, marshalErr)
	missingFieldsCursor := base64.RawURLEncoding.EncodeToString(payload)
	_, _, _, err = s.GetEventsForUser(ctx, user.ID, true, false, nil, nil, "", missingFieldsCursor, "UTC")
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidCursor)
}

func TestDashboardCursorEncodeDecode_RoundTripAndValidation(t *testing.T) {
	original := store.DashboardCursorForTest{
		EventDate: time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC),
		ID:        uuid.New(),
	}

	encoded, err := store.EncodeDashboardCursorForTest(original)
	require.NoError(t, err)
	require.NotEmpty(t, encoded)

	decoded, err := store.DecodeDashboardCursorForTest(encoded)
	require.NoError(t, err)
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.EventDate.UTC().Unix(), decoded.EventDate.UTC().Unix())

	_, err = store.DecodeDashboardCursorForTest("%%%")
	require.Error(t, err)

	badPayload, err := json.Marshal(map[string]any{
		"event_date": time.Now().UTC(),
		"id":         uuid.Nil,
	})
	require.NoError(t, err)
	_, err = store.DecodeDashboardCursorForTest(base64.RawURLEncoding.EncodeToString(badPayload))
	require.Error(t, err)
}

func TestGetEventGroupDetail_Success(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()

	host := createTestUser(t, ctx, s)
	participant := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)
	mode := firstModeForGame(t, ctx, s, games[0].ID)

	start := time.Date(2026, 6, 1, 15, 0, 0, 0, time.UTC)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, start)
	registerUserForEvent(t, ctx, tx, eventID, participant.ID)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	assert.Equal(t, groupID, detail.ID)
	assert.Equal(t, "balanced", detail.SortLogic)
	assert.Equal(t, host.ID, detail.OwnerID)
	assert.Equal(t, mode.ID, detail.GameModeID)
	assert.Greater(t, len(detail.Events), 0)
	require.Len(t, detail.Events, 1)
	assert.Equal(t, eventID, detail.Events[0].ID)
	assert.GreaterOrEqual(t, detail.Events[0].RegisteredCount, 1)
	require.GreaterOrEqual(t, len(detail.Events[0].Registrations), 1)
}

func TestGetEventGroupDetail_NotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	user := createTestUser(t, ctx, s)

	_, err := s.GetEventGroupDetail(ctx, uuid.New(), user.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventGroupNotFound)
}

func TestUpdateEventGroupSettings_Success(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, " EU-West ", 5, "ranked", true, mode.ID)
	require.NoError(t, err)
	var region string
	var subMin int32
	var sortLogic string
	var regOpen bool
	err = tx.QueryRow(ctx, `SELECT region, sub_min, sort_logic, registration_open FROM event_groups WHERE id = $1`, groupID).Scan(&region, &subMin, &sortLogic, &regOpen)
	require.NoError(t, err)
	assert.Equal(t, "EU-West", region)
	assert.Equal(t, int32(5), subMin)
	assert.Equal(t, "ranked", sortLogic)
	assert.True(t, regOpen)
}

func TestUpdateEventGroupSettings_Forbidden(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	other := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpdateEventGroupSettings(ctx, groupID, other.ID, "AMER", 0, "", true, mode.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestUpdateEventGroupSettings_Invalid(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "  ", 0, "", true, mode.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidSubMin)
}

func TestUpdateEventGroupSettings_InvalidNegativeSubMin(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", -1, "", true, mode.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidSubMin)
}

func TestUpdateEventGroupSettings_NotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)

	err := s.UpdateEventGroupSettings(ctx, uuid.New(), host.ID, "AMER", 0, "", false, uuid.Nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventGroupNotFound)
}

func TestUpdateEventGroupSettings_InvalidSortLogic(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "not-a-mode", true, mode.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidSortLogic)
}

func TestUpdateEventGroupSettings_PreservesSortLogicWhenEmpty(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	_, err = tx.Exec(ctx, `UPDATE event_groups SET sort_logic = 'ranked' WHERE id = $1`, groupID)
	require.NoError(t, err)

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "LATAM", 1, "", true, mode.ID)
	require.NoError(t, err)

	var region string
	var sortLogic string
	err = tx.QueryRow(ctx, `SELECT region, sort_logic FROM event_groups WHERE id = $1`, groupID).Scan(&region, &sortLogic)
	require.NoError(t, err)
	assert.Equal(t, "LATAM", region)
	assert.Equal(t, "ranked", sortLogic)
}

func TestUpdateEventGroupSettings_UpdatesGameModeWithinSameGame(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeA := firstModeForGame(t, ctx, s, games[0].ID)
	modeBID := insertExtraGameMode(t, ctx, tx, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, modeA.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, modeBID)
	require.NoError(t, err)
	var stored uuid.UUID
	err = tx.QueryRow(ctx, `SELECT game_mode_id FROM event_groups WHERE id = $1`, groupID).Scan(&stored)
	require.NoError(t, err)
	assert.Equal(t, modeBID, stored)
}

func TestUpdateEventGroupSettings_RejectsGameModeFromOtherGame(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(games), 2)
	modeGameA := firstModeForGame(t, ctx, s, games[0].ID)
	modeGameB := firstModeForGame(t, ctx, s, games[1].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, modeGameA.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, modeGameB.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrGameModeWrongGame)
}

func TestUpdateEventGroupSettings_NewGameModeNotFound(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrGameModeNotFound)
}

// TestUpdateEventGroupSettings_CurrentGameModeNotFound simulates a row whose game_mode_id no longer
// resolves (e.g. legacy/orphan data). A plain UPDATE to a missing mode_id violates the FK unless the
// constraint is deferred; this test uses DEFERRABLE so the violation is only checked at commit—cleanup rolls back.
func TestUpdateEventGroupSettings_CurrentGameModeNotFound(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	missingCurrentModeID := uuid.New()
	_, err = tx.Exec(ctx, `SET CONSTRAINTS event_groups_game_mode_id_fkey DEFERRED`)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `UPDATE event_groups SET game_mode_id = $1 WHERE id = $2`, missingCurrentModeID, groupID)
	require.NoError(t, err)

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, mode.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrGameModeNotFound)
}

func TestUpdateEventGroupSettings_SetsRegistrationClosed(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", false, mode.ID)
	require.NoError(t, err)
	var open bool
	err = tx.QueryRow(ctx, `SELECT registration_open FROM event_groups WHERE id = $1`, groupID).Scan(&open)
	require.NoError(t, err)
	assert.False(t, open)
}

func TestUpdateEventGroupSettings_OpenBlockedByExistingTeams(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	_, err = tx.Exec(ctx, `UPDATE event_groups SET registration_open = false WHERE id = $1`, groupID)
	require.NoError(t, err)
	insertLobbyForEvent(t, ctx, tx, eventID, &host.ID)

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, mode.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrOpenRegistrationTeams)
}

func TestDeleteEventGroup_Success(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.DeleteEventGroup(ctx, groupID, host.ID)
	require.NoError(t, err)
	var n int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM event_groups WHERE id = $1`, groupID).Scan(&n)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestDeleteEventGroup_Forbidden(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	other := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.DeleteEventGroup(ctx, groupID, other.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestDeleteEventGroup_NotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)

	err := s.DeleteEventGroup(ctx, uuid.New(), host.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, pgx.ErrNoRows))
}

func TestSetEventGroupRegistrationOpen_Success(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.SetEventGroupRegistrationOpen(ctx, groupID, host.ID, false)
	require.NoError(t, err)
	var open bool
	err = tx.QueryRow(ctx, `SELECT registration_open FROM event_groups WHERE id = $1`, groupID).Scan(&open)
	require.NoError(t, err)
	assert.False(t, open)

	err = s.SetEventGroupRegistrationOpen(ctx, groupID, host.ID, true)
	require.NoError(t, err)
	err = tx.QueryRow(ctx, `SELECT registration_open FROM event_groups WHERE id = $1`, groupID).Scan(&open)
	require.NoError(t, err)
	assert.True(t, open)
}

func TestSetEventGroupRegistrationOpen_Forbidden(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	other := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.SetEventGroupRegistrationOpen(ctx, groupID, other.ID, true)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestSetEventGroupRegistrationOpen_BlockedByExistingTeams(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	insertLobbyForEvent(t, ctx, tx, eventID, &host.ID)

	err = s.SetEventGroupRegistrationOpen(ctx, groupID, host.ID, true)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrOpenRegistrationTeams)
}

func TestSetEventGroupRegistrationOpen_NotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)

	err := s.SetEventGroupRegistrationOpen(ctx, uuid.New(), host.ID, true)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventGroupNotFound)
}

func TestCreateTeamsForGroup_Success(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	p1 := createTestUser(t, ctx, s)
	p2 := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	registerUserForEventWithFlags(t, ctx, tx, eventID, p1.ID, false, true)
	registerUserForEvent(t, ctx, tx, eventID, p2.ID)

	err = s.CreateTeamsForGroup(ctx, groupID, host.ID)
	require.NoError(t, err)
	var n int
	err = tx.QueryRow(ctx, `
		SELECT count(*) FROM lobbies L
		JOIN events E ON E.id = L.event_id
		WHERE E.group_id = $1`, groupID).Scan(&n)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	var regClosed bool
	err = tx.QueryRow(ctx, `SELECT registration_open FROM event_groups WHERE id = $1`, groupID).Scan(&regClosed)
	require.NoError(t, err)
	assert.False(t, regClosed)
}

func TestCreateTeamsForGroup_NoLobbyHostEligibleLeavesLobbyHostNull(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	p1 := createTestUser(t, ctx, s)
	p2 := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	registerUserForEventWithFlags(t, ctx, tx, eventID, p1.ID, true, false)
	registerUserForEventWithFlags(t, ctx, tx, eventID, p2.ID, false, false)

	err = s.CreateTeamsForGroup(ctx, groupID, host.ID)
	require.NoError(t, err)

	var lobbyHost *uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT L.host
		FROM lobbies L
		JOIN events E ON E.id = L.event_id
		WHERE E.group_id = $1
		LIMIT 1`, groupID).Scan(&lobbyHost)
	require.NoError(t, err)
	assert.Nil(t, lobbyHost)
}

func TestCreateTeamsForGroup_TeamsAlreadyCreated(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	p1 := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	registerUserForEvent(t, ctx, tx, eventID, p1.ID)
	insertLobbyForEvent(t, ctx, tx, eventID, nil)

	err = s.CreateTeamsForGroup(ctx, groupID, host.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrTeamsAlreadyCreated)
}

func TestCreateTeamsForGroup_Forbidden(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	other := createTestUser(t, ctx, s)
	p1 := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	registerUserForEvent(t, ctx, tx, eventID, p1.ID)

	err = s.CreateTeamsForGroup(ctx, groupID, other.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestCreateTeamsForGroup_NotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)

	err := s.CreateTeamsForGroup(ctx, uuid.New(), host.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventGroupNotFound)
}

func TestDeleteTeamsForGroup_Success(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	p1 := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	registerUserForEvent(t, ctx, tx, eventID, p1.ID)
	insertLobbyForEvent(t, ctx, tx, eventID, nil)
	_, err = tx.Exec(ctx, `UPDATE event_groups SET registration_open = false WHERE id = $1`, groupID)
	require.NoError(t, err)

	err = s.DeleteTeamsForGroup(ctx, groupID, host.ID)
	require.NoError(t, err)
	var n int
	err = tx.QueryRow(ctx, `
		SELECT count(*) FROM lobbies L
		JOIN events E ON E.id = L.event_id
		WHERE E.group_id = $1`, groupID).Scan(&n)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	var regOpen bool
	err = tx.QueryRow(ctx, `SELECT registration_open FROM event_groups WHERE id = $1`, groupID).Scan(&regOpen)
	require.NoError(t, err)
	assert.False(t, regOpen)
}

func TestDeleteTeamsForGroup_NoTeams(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, _ := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.DeleteTeamsForGroup(ctx, groupID, host.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrTeamsNotCreated)
}

func TestDeleteTeamsForGroup_Forbidden(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	other := createTestUser(t, ctx, s)
	p1 := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	registerUserForEvent(t, ctx, tx, eventID, p1.ID)
	insertLobbyForEvent(t, ctx, tx, eventID, nil)

	err = s.DeleteTeamsForGroup(ctx, groupID, other.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestDeleteTeamsForGroup_NotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)

	err := s.DeleteTeamsForGroup(ctx, uuid.New(), host.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventGroupNotFound)
}

func TestUpsertRegistrationForEvent_InsertAndUpdate(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	u := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, u.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	createCompleteUserGameForRegistration(t, ctx, tx, s, u.ID, games[0].ID)

	dr := "partner"
	err = s.UpsertRegistrationForEvent(ctx, eventID, u.ID, true, false, &dr)
	require.NoError(t, err)
	var canSub, canHost bool
	var duo *string
	err = tx.QueryRow(ctx, `
		SELECT can_substitute, can_lobby_host, duo_request FROM registrations WHERE event_id = $1 AND user_id = $2`,
		eventID, u.ID,
	).Scan(&canSub, &canHost, &duo)
	require.NoError(t, err)
	assert.True(t, canSub)
	assert.False(t, canHost)
	require.NotNil(t, duo)
	assert.Equal(t, "partner", *duo)

	err = s.UpsertRegistrationForEvent(ctx, eventID, u.ID, false, true, nil)
	require.NoError(t, err)
	err = tx.QueryRow(ctx, `
		SELECT can_substitute, can_lobby_host, duo_request FROM registrations WHERE event_id = $1 AND user_id = $2`,
		eventID, u.ID,
	).Scan(&canSub, &canHost, &duo)
	require.NoError(t, err)
	assert.False(t, canSub)
	assert.True(t, canHost)
	assert.Nil(t, duo)
}

func TestUpsertRegistrationForEvent_IncompleteUserGameProfile(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	joiner := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpsertRegistrationForEvent(ctx, eventID, joiner.ID, true, true, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrUserGameProfileIncomplete)
}

func TestUpsertRegistrationForEvent_RegistrationClosed(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	joiner := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	_, err = tx.Exec(ctx, `UPDATE event_groups SET registration_open = false WHERE id = (
		SELECT group_id FROM events WHERE id = $1
	)`, eventID)
	require.NoError(t, err)

	err = s.UpsertRegistrationForEvent(ctx, eventID, joiner.ID, true, true, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrRegistrationClosed)
}

func TestUpsertRegistrationForEvent_EventNotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	u := createTestUser(t, ctx, s)

	err := s.UpsertRegistrationForEvent(ctx, uuid.New(), u.ID, true, true, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventNotFound)
}

func TestUpsertRegistrationsForGroup_InsertUpdateDelete(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	participant := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	createCompleteUserGameForRegistration(t, ctx, tx, s, participant.ID, games[0].ID)

	groupID, event1 := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	event2 := insertEventInGroupFixture(t, ctx, tx, groupID, time.Now().UTC().Add(48*time.Hour))
	event3 := insertEventInGroupFixture(t, ctx, tx, groupID, time.Now().UTC().Add(72*time.Hour))

	registerUserForEventWithFlags(t, ctx, tx, event1, participant.ID, false, false)
	registerUserForEventWithFlags(t, ctx, tx, event2, participant.ID, true, false)

	duo := "carry-me"
	err = s.UpsertRegistrationsForGroup(ctx, groupID, participant.ID, []store.RegistrationUpsertItem{
		{EventID: event1, CanSubstitute: true, CanLobbyHost: true},  // update existing
		{EventID: event3, CanSubstitute: false, CanLobbyHost: true}, // insert new
	}, &duo)
	require.NoError(t, err)

	var canSub, canHost bool
	var duoRequest *string
	err = tx.QueryRow(ctx, `
		SELECT can_substitute, can_lobby_host, duo_request
		FROM registrations WHERE event_id = $1 AND user_id = $2`,
		event1, participant.ID).Scan(&canSub, &canHost, &duoRequest)
	require.NoError(t, err)
	assert.True(t, canSub)
	assert.True(t, canHost)
	require.NotNil(t, duoRequest)
	assert.Equal(t, duo, *duoRequest)

	err = tx.QueryRow(ctx, `
		SELECT can_substitute, can_lobby_host, duo_request
		FROM registrations WHERE event_id = $1 AND user_id = $2`,
		event3, participant.ID).Scan(&canSub, &canHost, &duoRequest)
	require.NoError(t, err)
	assert.False(t, canSub)
	assert.True(t, canHost)
	require.NotNil(t, duoRequest)
	assert.Equal(t, duo, *duoRequest)

	var deletedCount int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM registrations WHERE event_id = $1 AND user_id = $2`, event2, participant.ID).Scan(&deletedCount)
	require.NoError(t, err)
	assert.Equal(t, 0, deletedCount)
}

func TestUpsertRegistrationsForGroup_ErrorCases(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	participant := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpsertRegistrationsForGroup(ctx, groupID, participant.ID, nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidRegistration)

	err = s.UpsertRegistrationsForGroup(ctx, uuid.New(), participant.ID, []store.RegistrationUpsertItem{
		{EventID: eventID, CanSubstitute: true, CanLobbyHost: false},
	}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventGroupNotFound)

	_, err = tx.Exec(ctx, `UPDATE event_groups SET registration_open = false WHERE id = $1`, groupID)
	require.NoError(t, err)
	err = s.UpsertRegistrationsForGroup(ctx, groupID, participant.ID, []store.RegistrationUpsertItem{
		{EventID: eventID, CanSubstitute: true, CanLobbyHost: false},
	}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrRegistrationClosed)

	_, err = tx.Exec(ctx, `UPDATE event_groups SET registration_open = true WHERE id = $1`, groupID)
	require.NoError(t, err)
	createCompleteUserGameForRegistration(t, ctx, tx, s, participant.ID, games[0].ID)
	err = s.UpsertRegistrationsForGroup(ctx, groupID, participant.ID, []store.RegistrationUpsertItem{
		{EventID: uuid.New(), CanSubstitute: true, CanLobbyHost: false},
	}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventNotFound)
}

func TestUpsertRegistrationsForGroup_IncompleteUserGameProfile(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	participant := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.UpsertRegistrationsForGroup(ctx, groupID, participant.ID, []store.RegistrationUpsertItem{
		{EventID: eventID, CanSubstitute: true, CanLobbyHost: false},
	}, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrUserGameProfileIncomplete)
}

func TestDeleteRegistrationForEvent_SelfWhileOpen(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	joiner := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	registerUserForEvent(t, ctx, tx, eventID, joiner.ID)

	err = s.DeleteRegistrationForEvent(ctx, eventID, joiner.ID, joiner.ID)
	require.NoError(t, err)
	var n int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM registrations WHERE event_id = $1 AND user_id = $2`, eventID, joiner.ID).Scan(&n)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestDeleteRegistrationForEvent_HostDeletesParticipantWhenClosed(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	joiner := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	registerUserForEvent(t, ctx, tx, eventID, joiner.ID)
	_, err = tx.Exec(ctx, `UPDATE event_groups SET registration_open = false WHERE id = (
		SELECT group_id FROM events WHERE id = $1
	)`, eventID)
	require.NoError(t, err)

	err = s.DeleteRegistrationForEvent(ctx, eventID, joiner.ID, host.ID)
	require.NoError(t, err)
}

func TestDeleteRegistrationForEvent_ForbiddenNonHostDeletingOther(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	u1 := createTestUser(t, ctx, s)
	u2 := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	registerUserForEvent(t, ctx, tx, eventID, u1.ID)
	registerUserForEvent(t, ctx, tx, eventID, u2.ID)

	err = s.DeleteRegistrationForEvent(ctx, eventID, u2.ID, u1.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestDeleteRegistrationForEvent_SelfWhenClosed(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	joiner := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))
	registerUserForEvent(t, ctx, tx, eventID, joiner.ID)
	_, err = tx.Exec(ctx, `UPDATE event_groups SET registration_open = false WHERE id = (
		SELECT group_id FROM events WHERE id = $1
	)`, eventID)
	require.NoError(t, err)

	err = s.DeleteRegistrationForEvent(ctx, eventID, joiner.ID, joiner.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrRegistrationClosed)
}

func TestDeleteRegistrationForEvent_NotFound(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	joiner := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, mode.ID, time.Now().UTC().Add(24*time.Hour))

	err = s.DeleteRegistrationForEvent(ctx, eventID, joiner.ID, joiner.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrRegistrationNotFound)
}

func TestDeleteRegistrationForEvent_EventNotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	u := createTestUser(t, ctx, s)

	err := s.DeleteRegistrationForEvent(ctx, uuid.New(), u.ID, u.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventNotFound)
}
