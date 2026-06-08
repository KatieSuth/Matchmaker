package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntPtrToInt32PtrForTest(t *testing.T) {
	assert.Nil(t, store.IntPtrToInt32PtrForTest(nil))

	teamNum := 2
	out := store.IntPtrToInt32PtrForTest(&teamNum)
	require.NotNil(t, out)
	assert.Equal(t, int32(2), *out)
}

func TestPersistTeamPlansForTest_PersistsRosterSubsAndFairness(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	p1 := createTestUser(t, ctx, s)
	p2 := createTestUser(t, ctx, s)
	sub := createTestUser(t, ctx, s)
	team1 := 1
	team2 := 2

	err = store.PersistTeamPlansForTest(s, ctx, []matchmaking.GamePlan{{
		EventID: eventID,
		Lobbies: []matchmaking.LobbyPlan{{
			HostID:          &p1.ID,
			FairnessWarning: true,
			Roster: []matchmaking.Player{
				{UserID: p1.ID, TeamNumber: &team1},
				{UserID: p2.ID, TeamNumber: &team2},
			},
			Subs: []matchmaking.Player{
				{UserID: sub.ID},
			},
		}},
	}})
	require.NoError(t, err)

	var lobbyCount, playerCount int
	var fairnessWarning bool
	var subCount int
	err = tx.QueryRow(ctx, `
		SELECT count(*), bool_or(fairness_warning)
		FROM lobbies
		WHERE event_id = $1`, eventID).Scan(&lobbyCount, &fairnessWarning)
	require.NoError(t, err)
	assert.Equal(t, 1, lobbyCount)
	assert.True(t, fairnessWarning)

	err = tx.QueryRow(ctx, `
		SELECT count(*)
		FROM players P
		JOIN lobbies L ON L.id = P.lobby_id
		WHERE L.event_id = $1`, eventID).Scan(&playerCount)
	require.NoError(t, err)
	assert.Equal(t, 3, playerCount)

	err = tx.QueryRow(ctx, `
		SELECT count(*)
		FROM players P
		JOIN lobbies L ON L.id = P.lobby_id
		WHERE L.event_id = $1 AND P.team_number IS NULL`, eventID).Scan(&subCount)
	require.NoError(t, err)
	assert.Equal(t, 1, subCount)
}

func TestPersistTeamPlansForTest_CreateLobbyError(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()

	err := store.PersistTeamPlansForTest(s, ctx, []matchmaking.GamePlan{{
		EventID: uuid.New(),
		Lobbies: []matchmaking.LobbyPlan{{}},
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create lobby")
}

func TestPersistTeamPlansForTest_CreateRosterPlayerError(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	team1 := 1
	err = store.PersistTeamPlansForTest(s, ctx, []matchmaking.GamePlan{{
		EventID: eventID,
		Lobbies: []matchmaking.LobbyPlan{{
			Roster: []matchmaking.Player{
				{UserID: uuid.New(), TeamNumber: &team1},
			},
		}},
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create roster player")
}

func TestPersistTeamPlansForTest_CreateSubPlayerError(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	p1 := createTestUser(t, ctx, s)
	team1 := 1

	err = store.PersistTeamPlansForTest(s, ctx, []matchmaking.GamePlan{{
		EventID: eventID,
		Lobbies: []matchmaking.LobbyPlan{{
			Roster: []matchmaking.Player{
				{UserID: p1.ID, TeamNumber: &team1},
			},
			Subs: []matchmaking.Player{
				{UserID: uuid.New()},
			},
		}},
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create sub player")
}

func TestMapRegistrationsToPlayersForTest_MapsDuoFields(t *testing.T) {
	userID := uuid.New()
	duo := "partner"
	name := "alpha"
	now := time.Now()

	players := store.MapRegistrationsToPlayersForTest([]db.GetMatchmakingRegistrationsForEventRow{
		{
			UserID:           userID,
			CanSubstitute:    true,
			CanLobbyHost:     false,
			DuoRequest:       &duo,
			CreatedAt:        now,
			DiscordName:      &name,
			CurrentRankOrder: 10,
			PeakRankOrder:    12,
		},
	}, map[uuid.UUID]int{userID: 2})

	require.Len(t, players, 1)
	assert.Equal(t, "alpha", players[0].DiscordName)
	require.NotNil(t, players[0].DuoRequest)
	assert.Equal(t, "partner", *players[0].DuoRequest)
	assert.Equal(t, 2, players[0].RegisteredGameCount)
}

func TestMapRegistrationsToPlayersForTest_NilOptionalFields(t *testing.T) {
	userID := uuid.New()
	players := store.MapRegistrationsToPlayersForTest([]db.GetMatchmakingRegistrationsForEventRow{
		{
			UserID:           userID,
			CanSubstitute:    false,
			CanLobbyHost:     false,
			DuoRequest:       nil,
			CreatedAt:        time.Now(),
			DiscordName:      nil,
			CurrentRankOrder: 4,
			PeakRankOrder:    4,
		},
	}, map[uuid.UUID]int{userID: 1})

	require.Len(t, players, 1)
	assert.Empty(t, players[0].DiscordName)
	assert.Nil(t, players[0].DuoRequest)
}
