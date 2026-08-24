package store_test

import (
	"context"
	"errors"
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

type faultInjectTx struct {
	db.DBTX
	queryRowCalls      int
	failOnQueryRowCall int
	queryCalls         int
	failOnQueryCall    int
	execCalls          int
	failOnExecCall     int
	injectedErr        error
}

func (f *faultInjectTx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	f.queryRowCalls++
	if f.queryRowCalls == f.failOnQueryRowCall {
		return faultRow{err: f.injectedErr}
	}
	return f.DBTX.QueryRow(ctx, sql, args...)
}

func (f *faultInjectTx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	f.queryCalls++
	if f.queryCalls == f.failOnQueryCall {
		return nil, f.injectedErr
	}
	return f.DBTX.Query(ctx, sql, args...)
}

func (f *faultInjectTx) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	f.execCalls++
	if f.execCalls == f.failOnExecCall {
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
		name               string
		failOnQueryRowCall int
		failOnQueryCall    int
		failOnExecCall     int
		wantContains       string
	}{
		{
			name:               "count lobbies fails",
			failOnQueryRowCall: 2,
			wantContains:       "count lobbies for event",
		},
		{
			name:               "registration lookup fails",
			failOnQueryRowCall: 3,
			wantContains:       "injected database failure",
		},
		{
			name:            "player placements fails",
			failOnQueryCall: 1,
			wantContains:    "get player placements",
		},
		{
			name:            "lobbies lookup fails",
			failOnQueryCall: 2,
			wantContains:    "get lobbies for event",
		},
		{
			name:           "update lobby host fails",
			failOnExecCall: 1,
			wantContains:   "update lobby host",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupLobbyHostFixture(t)
			faulty := &faultInjectTx{
				DBTX:               fixture.tx,
				failOnQueryRowCall: tc.failOnQueryRowCall,
				failOnQueryCall:    tc.failOnQueryCall,
				failOnExecCall:     tc.failOnExecCall,
				injectedErr:        injectedErr,
			}
			s := store.NewPostgresStoreFromDBTXForTest(faulty)

			err := s.SetLobbyHostForEvent(fixture.ctx, fixture.eventID, fixture.ownerID, fixture.userID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantContains)
		})
	}
}