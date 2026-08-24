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

type swapFixtureData struct {
	tx      db.DBTX
	ctx     context.Context
	eventID uuid.UUID
	ownerID uuid.UUID
	userA   uuid.UUID
	userB   uuid.UUID
}

func setupSwapFixture(t *testing.T) swapFixtureData {
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
	p1, _, ok := findLobbyTeamPlayer(detail, eventID, 1)
	require.True(t, ok)
	p2, _, ok := findLobbyTeamPlayer(detail, eventID, 2)
	require.True(t, ok)

	return swapFixtureData{
		tx:      tx,
		ctx:     ctx,
		eventID: eventID,
		ownerID: host.ID,
		userA:   p1.UserID,
		userB:   p2.UserID,
	}
}

func TestSwapPlayersForEvent_DatabaseErrors(t *testing.T) {
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
			name:           "delete player A fails",
			failOnExecCall: 1,
			wantContains:   "delete player A",
		},
		{
			name:           "delete player B fails",
			failOnExecCall: 2,
			wantContains:   "delete player B",
		},
		{
			name:           "insert player A fails",
			failOnExecCall: 3,
			wantContains:   "insert player A",
		},
		{
			name:           "insert player B fails",
			failOnExecCall: 4,
			wantContains:   "insert player B",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupSwapFixture(t)
			faulty := &faultInjectTx{
				DBTX:               fixture.tx,
				failOnQueryRowCall: tc.failOnQueryRowCall,
				failOnQueryCall:    tc.failOnQueryCall,
				failOnExecCall:     tc.failOnExecCall,
				injectedErr:        injectedErr,
			}
			s := store.NewPostgresStoreFromDBTXForTest(faulty)

			err := s.SwapPlayersForEvent(
				fixture.ctx,
				fixture.eventID,
				fixture.ownerID,
				fixture.userA,
				fixture.userB,
				defaultMatchmakingSettings(),
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantContains)
		})
	}
}

func TestRecomputeLobbyAfterSwap_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))
	lobbyID := insertLobbyForEvent(t, ctx, tx, eventID, &host.ID)

	cases := []struct {
		name               string
		failOnQueryCall    int
		failOnQueryRowCall int
		failOnExecCall     int
		wantContains       string
	}{
		{
			name:            "get players fails",
			failOnQueryCall: 1,
			wantContains:    "get players for lobby",
		},
		{
			name:           "update lobby host fails",
			failOnExecCall: 1,
			wantContains:   "update lobby host",
		},
		{
			name:               "get max rank order fails",
			failOnQueryRowCall: 2,
			wantContains:       "get max rank order",
		},
		{
			name:           "update fairness warning fails",
			failOnExecCall: 2,
			wantContains:   "update lobby fairness",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			faulty := &faultInjectTx{
				DBTX:               tx,
				failOnQueryCall:    tc.failOnQueryCall,
				failOnQueryRowCall: tc.failOnQueryRowCall,
				failOnExecCall:     tc.failOnExecCall,
				injectedErr:        injectedErr,
			}
			fs := store.NewPostgresStoreFromDBTXForTest(faulty)

			err := store.RecomputeLobbyAfterSwapForTest(fs, ctx, lobbyID, defaultMatchmakingSettings(), modeID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantContains)
		})
	}
}
