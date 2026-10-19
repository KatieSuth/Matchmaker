package store_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// faultInjectTx wraps a DBTX and fails one named sqlc query (the "-- name: X" header sqlc puts on
// every generated query), so fault tests survive queries being added or reordered.
type faultInjectTx struct {
	db.DBTX
	// failOn is the sqlc query name to fail, e.g. "CountLobbiesForEvent"; empty never fails.
	failOn string
	// failOnNth picks which call of failOn fails (0 or 1 = the first).
	failOnNth   int
	injectedErr error
	calls       map[string]int
	totalCalls  int
}

var sqlcQueryName = regexp.MustCompile(`-- name: (\w+)`)

// shouldFail counts the call and reports whether it is the configured failure.
func (f *faultInjectTx) shouldFail(sql string) bool {
	f.totalCalls++
	name := ""
	if m := sqlcQueryName.FindStringSubmatch(sql); m != nil {
		name = m[1]
	}
	if f.calls == nil {
		f.calls = map[string]int{}
	}
	f.calls[name]++
	nth := f.failOnNth
	if nth < 1 {
		nth = 1
	}
	return f.failOn != "" && name == f.failOn && f.calls[name] == nth
}

func (f *faultInjectTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if f.shouldFail(sql) {
		return faultRow{err: f.injectedErr}
	}
	return f.DBTX.QueryRow(ctx, sql, args...)
}

func (f *faultInjectTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if f.shouldFail(sql) {
		return nil, f.injectedErr
	}
	return f.DBTX.Query(ctx, sql, args...)
}

func (f *faultInjectTx) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if f.shouldFail(sql) {
		return pgconn.CommandTag{}, f.injectedErr
	}
	return f.DBTX.Exec(ctx, sql, args...)
}

type faultRow struct {
	err error
}

func (r faultRow) Scan(dest ...any) error {
	return r.err
}

type lobbyHostFixtureData struct {
	tx      db.DBTX
	ctx     context.Context
	eventID uuid.UUID
	ownerID uuid.UUID
	userID  uuid.UUID
}

func setupLobbyHostFixture(t *testing.T) lobbyHostFixtureData {
	t.Helper()
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	for i := 0; i < 4; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, false)
	}

	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	target, _, ok := findNonHostTeamPlayerInLobbyIndex(detail, eventID, 0)
	require.True(t, ok)

	return lobbyHostFixtureData{
		tx:      tx,
		ctx:     ctx,
		eventID: eventID,
		ownerID: host.ID,
		userID:  target.UserID,
	}
}

func TestSetLobbyHostForEvent_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")

	cases := []struct {
		name         string
		failOn       string
		failOnNth    int
		wantContains string
	}{
		{
			name:         "count lobbies fails",
			failOn:       "CountLobbiesForEvent",
			wantContains: "count lobbies for event",
		},
		{
			name:         "registration lookup fails",
			failOn:       "GetRegistrationByEventAndUser",
			wantContains: "injected database failure",
		},
		{
			name:         "player placements fails",
			failOn:       "GetPlayerPlacementsForEvent",
			wantContains: "get player placements",
		},
		{
			name:         "lobbies lookup fails",
			failOn:       "GetLobbiesForEvent",
			wantContains: "get lobbies for event",
		},
		{
			name:         "update lobby host fails",
			failOn:       "UpdateLobbyHost",
			wantContains: "update lobby host",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupLobbyHostFixture(t)
			faulty := &faultInjectTx{
				DBTX:        fixture.tx,
				failOn:      tc.failOn,
				failOnNth:   tc.failOnNth,
				injectedErr: injectedErr,
			}
			s := store.NewPostgresStoreFromDBTXForTest(faulty)

			err := s.SetLobbyHostForEvent(fixture.ctx, fixture.eventID, fixture.ownerID, fixture.userID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantContains)
		})
	}
}
