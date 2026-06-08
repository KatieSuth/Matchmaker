package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type moveSubFixtureData struct {
	tx      db.DBTX
	ctx     context.Context
	eventID uuid.UUID
	ownerID uuid.UUID
	subUser uuid.UUID
}

type moveUnplacedToSubsFixtureData struct {
	tx         db.DBTX
	ctx        context.Context
	eventID    uuid.UUID
	ownerID    uuid.UUID
	unplacedID uuid.UUID
	lobbyID    uuid.UUID
}

func setupMoveSubToUnplacedFixture(t *testing.T) moveSubFixtureData {
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
	require.NoError(t, s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings()))

	lobbyID := firstLobbyID(t, ctx, tx, eventID)
	subUser := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subUser.ID, games[0].ID, true, false)
	insertPlayerForLobby(t, ctx, tx, lobbyID, subUser.ID, nil)

	return moveSubFixtureData{
		tx:      tx,
		ctx:     ctx,
		eventID: eventID,
		ownerID: host.ID,
		subUser: subUser.ID,
	}
}

func setupMoveUnplacedToSubsFixture(t *testing.T) moveUnplacedToSubsFixtureData {
	t.Helper()
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	for i := 0; i < 5; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, false)
	}
	require.NoError(t, s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings()))

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	require.Len(t, detail.Events[0].Unplaced, 1)
	unplacedID := detail.Events[0].Unplaced[0].UserID
	_, err = tx.Exec(ctx, `
		UPDATE registrations SET can_substitute = true
		WHERE event_id = $1 AND user_id = $2`, eventID, unplacedID)
	require.NoError(t, err)

	return moveUnplacedToSubsFixtureData{
		tx:         tx,
		ctx:        ctx,
		eventID:    eventID,
		ownerID:    host.ID,
		unplacedID: unplacedID,
		lobbyID:    detail.Events[0].Lobbies[0].ID,
	}
}

func TestMoveSubToUnplacedForEvent_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")

	cases := []struct {
		name               string
		failOnQueryRowCall int
		failOnQueryCall    int
		failOnExecCall     int
		wantContains       string
	}{
		{
			name:               "event meta lookup fails",
			failOnQueryRowCall: 1,
			wantContains:       "get event group meta",
		},
		{
			name:               "count lobbies fails",
			failOnQueryRowCall: 2,
			wantContains:       "count lobbies for event",
		},
		{
			name:               "registration lookup fails",
			failOnQueryRowCall: 3,
			wantContains:       "get registration",
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
			name:           "delete substitute fails",
			failOnExecCall: 1,
			wantContains:   "delete substitute player",
		},
		{
			name:            "recompute lobby fails",
			failOnQueryCall: 3,
			wantContains:    "get players for lobby",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupMoveSubToUnplacedFixture(t)
			faulty := &faultInjectTx{
				DBTX:               fixture.tx,
				failOnQueryRowCall: tc.failOnQueryRowCall,
				failOnQueryCall:    tc.failOnQueryCall,
				failOnExecCall:     tc.failOnExecCall,
				injectedErr:        injectedErr,
			}
			s := store.NewPostgresStoreFromDBTXForTest(faulty)

			err := s.MoveSubToUnplacedForEvent(
				fixture.ctx,
				fixture.eventID,
				fixture.ownerID,
				fixture.subUser,
				defaultMatchmakingSettings(),
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantContains)
		})
	}
}

func TestMoveUnplacedToSubsForEvent_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")

	cases := []struct {
		name               string
		failOnQueryRowCall int
		failOnQueryCall    int
		failOnExecCall     int
		wantContains       string
	}{
		{
			name:               "event meta lookup fails",
			failOnQueryRowCall: 1,
			wantContains:       "get event group meta",
		},
		{
			name:               "count lobbies fails",
			failOnQueryRowCall: 2,
			wantContains:       "count lobbies for event",
		},
		{
			name:               "registration lookup fails",
			failOnQueryRowCall: 3,
			wantContains:       "get registration",
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
			name:           "insert substitute fails",
			failOnExecCall: 1,
			wantContains:   "insert substitute player",
		},
		{
			name:            "recompute lobby fails",
			failOnQueryCall: 3,
			wantContains:    "get players for lobby",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupMoveUnplacedToSubsFixture(t)
			faulty := &faultInjectTx{
				DBTX:               fixture.tx,
				failOnQueryRowCall: tc.failOnQueryRowCall,
				failOnQueryCall:    tc.failOnQueryCall,
				failOnExecCall:     tc.failOnExecCall,
				injectedErr:        injectedErr,
			}
			s := store.NewPostgresStoreFromDBTXForTest(faulty)

			err := s.MoveUnplacedToSubsForEvent(
				fixture.ctx,
				fixture.eventID,
				fixture.ownerID,
				fixture.unplacedID,
				fixture.lobbyID,
				defaultMatchmakingSettings(),
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantContains)
		})
	}
}
