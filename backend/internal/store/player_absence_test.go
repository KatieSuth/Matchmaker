package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type absenceFixture struct {
	s       *store.PostgresStore
	tx      db.DBTX
	ctx     context.Context
	groupID uuid.UUID
	eventID uuid.UUID
	hostID  uuid.UUID
	gameID  uuid.UUID
	target  uuid.UUID
}

// setupAbsenceFixture creates a locked-in game that started an hour ago, with four team players,
// and returns one as the target.
func setupAbsenceFixture(t *testing.T) absenceFixture {
	t.Helper()
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(-time.Hour))

	for i := 0; i < 4; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, false)
	}

	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	target, _, ok := findLobbyTeamPlayer(detail, eventID, 1)
	require.True(t, ok)

	return absenceFixture{
		s:       s,
		tx:      tx,
		ctx:     ctx,
		groupID: groupID,
		eventID: eventID,
		hostID:  host.ID,
		gameID:  games[0].ID,
		target:  target.UserID,
	}
}

// noShowRow reads the stored no-show for a player, reporting whether it exists.
func noShowRow(t *testing.T, f absenceFixture, userID uuid.UUID) (hostID uuid.UUID, markedAt time.Time, found bool) {
	t.Helper()
	var count int
	require.NoError(t, f.tx.QueryRow(f.ctx, `SELECT COUNT(*) FROM no_shows WHERE event_id = $1 AND user_id = $2`, f.eventID, userID).Scan(&count))
	if count == 0 {
		return uuid.Nil, time.Time{}, false
	}
	require.NoError(t, f.tx.QueryRow(f.ctx,
		`SELECT host_id, marked_absent_at FROM no_shows WHERE event_id = $1 AND user_id = $2`,
		f.eventID, userID,
	).Scan(&hostID, &markedAt))
	return hostID, markedAt, true
}

// absentIDsForEvent returns the absent_user_ids the given viewer sees for the fixture's game.
func absentIDsForEvent(t *testing.T, f absenceFixture, viewerID uuid.UUID) []uuid.UUID {
	t.Helper()
	detail, err := f.s.GetEventGroupDetail(f.ctx, f.groupID, viewerID)
	require.NoError(t, err)
	for _, event := range detail.Events {
		if event.ID == f.eventID {
			return event.AbsentUserIDs
		}
	}
	t.Fatalf("event %s not found in detail", f.eventID)
	return nil
}

func TestAbsenceValidationError_ErrorAndUnwrap(t *testing.T) {
	err := &store.AbsenceValidationError{Message: "Only players on a team can be marked absent"}
	assert.Equal(t, "Only players on a team can be marked absent", err.Error())
	assert.ErrorIs(t, err, store.ErrInvalidAbsenceChange)
}

func TestSetPlayerAbsenceForEvent_MarkAndClear(t *testing.T) {
	f := setupAbsenceFixture(t)

	require.NoError(t, f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, true))
	hostID, _, found := noShowRow(t, f, f.target)
	require.True(t, found)
	assert.Equal(t, f.hostID, hostID)

	require.NoError(t, f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, false))
	_, _, found = noShowRow(t, f, f.target)
	assert.False(t, found)
}

func TestSetPlayerAbsenceForEvent_RemarkKeepsOriginalTime(t *testing.T) {
	f := setupAbsenceFixture(t)
	original := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
	_, err := f.tx.Exec(f.ctx,
		`INSERT INTO no_shows (event_id, user_id, host_id, marked_absent_at) VALUES ($1, $2, $3, $4)`,
		f.eventID, f.target, f.hostID, original,
	)
	require.NoError(t, err)

	require.NoError(t, f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, true))
	_, markedAt, found := noShowRow(t, f, f.target)
	require.True(t, found)
	assert.True(t, original.Equal(markedAt.UTC()))
}

func TestSetPlayerAbsenceForEvent_ClearWhenNotMarkedIsNoOp(t *testing.T) {
	f := setupAbsenceFixture(t)
	require.NoError(t, f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, false))
}

func TestSetPlayerAbsenceForEvent_Forbidden(t *testing.T) {
	f := setupAbsenceFixture(t)
	other := createTestUser(t, f.ctx, f.s)

	err := f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, other.ID, f.target, true)
	assert.ErrorIs(t, err, store.ErrForbidden)
	_, _, found := noShowRow(t, f, f.target)
	assert.False(t, found, "non-host mark must not write a row")

	require.NoError(t, f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, true))
	err = f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, other.ID, f.target, false)
	assert.ErrorIs(t, err, store.ErrForbidden)
	_, _, found = noShowRow(t, f, f.target)
	assert.True(t, found, "non-host clear must not delete the host's mark")
}

func TestSetPlayerAbsenceForEvent_LobbyHostForbidden(t *testing.T) {
	f := setupAbsenceFixture(t)
	var lobbyHost uuid.UUID
	require.NoError(t, f.tx.QueryRow(f.ctx, `SELECT host FROM lobbies WHERE event_id = $1 AND host IS NOT NULL LIMIT 1`, f.eventID).Scan(&lobbyHost))
	require.NotEqual(t, f.hostID, lobbyHost)

	err := f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, lobbyHost, f.target, true)
	assert.ErrorIs(t, err, store.ErrForbidden)
	_, _, found := noShowRow(t, f, f.target)
	assert.False(t, found)
}

func TestSetPlayerAbsenceForEvent_PlayerCannotMarkThemselves(t *testing.T) {
	f := setupAbsenceFixture(t)
	err := f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.target, f.target, true)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestSetPlayerAbsenceForEvent_TeamsNotCreated(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	err = s.SetPlayerAbsenceForEvent(ctx, eventID, host.ID, uuid.New(), true)
	assert.ErrorIs(t, err, store.ErrTeamsNotCreated)
}

func TestSetPlayerAbsenceForEvent_EventNotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)

	err := s.SetPlayerAbsenceForEvent(ctx, uuid.New(), host.ID, uuid.New(), true)
	assert.ErrorIs(t, err, store.ErrEventNotFound)
}

func TestSetPlayerAbsenceForEvent_ContextCancelled(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.SetPlayerAbsenceForEvent(ctx, uuid.New(), uuid.New(), uuid.New(), true)
	require.Error(t, err)
}

func TestSetPlayerAbsenceForEvent_SubRejected(t *testing.T) {
	f := setupAbsenceFixture(t)
	sub := createTestUser(t, f.ctx, f.s)
	registerPlayerForEventWithProfile(t, f.ctx, f.tx, f.s, f.eventID, sub.ID, f.gameID, true, false)
	insertPlayerForLobby(t, f.ctx, f.tx, firstLobbyID(t, f.ctx, f.tx, f.eventID), sub.ID, nil)

	err := f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, sub.ID, true)
	assert.ErrorIs(t, err, store.ErrInvalidAbsenceChange)
	var absenceErr *store.AbsenceValidationError
	require.ErrorAs(t, err, &absenceErr)
	assert.Equal(t, "Only players on a team can be marked absent", absenceErr.Message)
}

func TestSetPlayerAbsenceForEvent_UnplacedRejected(t *testing.T) {
	f := setupAbsenceFixture(t)
	unplaced := createTestUser(t, f.ctx, f.s)
	registerPlayerForEventWithProfile(t, f.ctx, f.tx, f.s, f.eventID, unplaced.ID, f.gameID, false, false)

	err := f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, unplaced.ID, true)
	assert.ErrorIs(t, err, store.ErrInvalidAbsenceChange)
}

func TestSetPlayerAbsenceForEvent_MarkPersistsAfterSwapToSubs(t *testing.T) {
	f := setupAbsenceFixture(t)
	sub := createTestUser(t, f.ctx, f.s)
	registerPlayerForEventWithProfile(t, f.ctx, f.tx, f.s, f.eventID, sub.ID, f.gameID, true, false)
	insertPlayerForLobby(t, f.ctx, f.tx, firstLobbyID(t, f.ctx, f.tx, f.eventID), sub.ID, nil)

	require.NoError(t, f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, true))
	require.NoError(t, f.s.SwapPlayersForEvent(f.ctx, f.eventID, f.hostID, f.target, sub.ID, defaultMatchmakingSettings()))

	_, _, found := noShowRow(t, f, f.target)
	assert.True(t, found)
	assert.Equal(t, []uuid.UUID{f.target}, absentIDsForEvent(t, f, f.hostID))

	// Clearing still works once the player is no longer on a team.
	require.NoError(t, f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, false))
	_, _, found = noShowRow(t, f, f.target)
	assert.False(t, found)
}

func TestSetPlayerAbsenceForEvent_MarkPersistsAfterTeamsDeleted(t *testing.T) {
	f := setupAbsenceFixture(t)
	require.NoError(t, f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, true))
	require.NoError(t, f.s.DeleteTeamsForGroup(f.ctx, f.groupID, f.hostID))

	_, _, found := noShowRow(t, f, f.target)
	assert.True(t, found)
	assert.Equal(t, []uuid.UUID{f.target}, absentIDsForEvent(t, f, f.hostID))
}

func TestGetEventGroupDetail_AbsentUserIDsVisibility(t *testing.T) {
	f := setupAbsenceFixture(t)
	require.NoError(t, f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, true))
	bystander := createTestUser(t, f.ctx, f.s)

	assert.Equal(t, []uuid.UUID{f.target}, absentIDsForEvent(t, f, f.hostID))
	assert.Equal(t, []uuid.UUID{f.target}, absentIDsForEvent(t, f, f.target))
	got := absentIDsForEvent(t, f, bystander.ID)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestVisibleAbsentUserIDs(t *testing.T) {
	viewer := uuid.New()
	other := uuid.New()

	cases := []struct {
		name         string
		absentIDs    []uuid.UUID
		viewerIsHost bool
		want         []uuid.UUID
	}{
		{name: "host sees all", absentIDs: []uuid.UUID{viewer, other}, viewerIsHost: true, want: []uuid.UUID{viewer, other}},
		{name: "host nil list", absentIDs: nil, viewerIsHost: true, want: []uuid.UUID{}},
		{name: "player sees own mark", absentIDs: []uuid.UUID{other, viewer}, want: []uuid.UUID{viewer}},
		{name: "player not marked", absentIDs: []uuid.UUID{other}, want: []uuid.UUID{}},
		{name: "player nil list", absentIDs: nil, want: []uuid.UUID{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := store.VisibleAbsentUserIDsForTest(tc.absentIDs, viewer, tc.viewerIsHost)
			require.NotNil(t, got)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSetPlayerAbsenceForEvent_RejectsMarkBeforeGameStarts(t *testing.T) {
	f := setupAbsenceFixture(t)
	beforeStart := time.Now().Add(-2 * time.Hour)

	err := store.SetPlayerAbsenceForEventAtForTest(f.s, f.ctx, f.eventID, f.hostID, f.target, true, beforeStart)
	assert.ErrorIs(t, err, store.ErrInvalidAbsenceChange)
	var absenceErr *store.AbsenceValidationError
	require.ErrorAs(t, err, &absenceErr)
	assert.Equal(t, "Players can be marked absent once the game has started", absenceErr.Message)
	_, _, found := noShowRow(t, f, f.target)
	assert.False(t, found)
}

func TestSetPlayerAbsenceForEvent_ClearAllowedBeforeGameStarts(t *testing.T) {
	f := setupAbsenceFixture(t)
	require.NoError(t, f.s.SetPlayerAbsenceForEvent(f.ctx, f.eventID, f.hostID, f.target, true))

	require.NoError(t, store.SetPlayerAbsenceForEventAtForTest(f.s, f.ctx, f.eventID, f.hostID, f.target, false, time.Now().Add(-2*time.Hour)))
	_, _, found := noShowRow(t, f, f.target)
	assert.False(t, found)
}
