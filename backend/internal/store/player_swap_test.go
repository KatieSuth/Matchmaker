package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findLobbyTeamPlayer(detail model.EventGroupDetail, eventID uuid.UUID, teamNumber int) (model.LobbyPlayer, uuid.UUID, bool) {
	for _, event := range detail.Events {
		if event.ID != eventID {
			continue
		}
		for _, lobby := range event.Lobbies {
			for _, team := range lobby.Teams {
				if team.TeamNumber != teamNumber {
					continue
				}
				if len(team.Players) == 0 {
					continue
				}
				return team.Players[0], lobby.ID, true
			}
		}
	}
	return model.LobbyPlayer{}, uuid.Nil, false
}

func insertPlayerForLobby(t *testing.T, ctx context.Context, tx db.DBTX, lobbyID, userID uuid.UUID, teamNumber *int32) {
	t.Helper()
	_, err := tx.Exec(ctx, `
		INSERT INTO players (lobby_id, user_id, team_number, event_id, created_at, updated_at)
		VALUES ($1, $2, $3, (SELECT event_id FROM lobbies WHERE id = $1), NOW(), NOW())
	`, lobbyID, userID, teamNumber)
	require.NoError(t, err)
}

func firstLobbyID(t *testing.T, ctx context.Context, tx db.DBTX, eventID uuid.UUID) uuid.UUID {
	t.Helper()
	var lobbyID uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM lobbies WHERE event_id = $1 ORDER BY created_at ASC LIMIT 1`, eventID).Scan(&lobbyID)
	require.NoError(t, err)
	return lobbyID
}

func teamNumberPtr(n int32) *int32 {
	return &n
}

func TestSwapValidationError_ErrorAndUnwrap(t *testing.T) {
	err := &store.SwapValidationError{Message: "Cannot swap players on the same team"}
	assert.Equal(t, "Cannot swap players on the same team", err.Error())
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
}

func TestSameTeamPlacementForTest(t *testing.T) {
	lobbyID := uuid.New()
	team1 := teamNumberPtr(1)

	assert.False(t, store.SameTeamPlacementForTest(
		store.SwapPlacementForTest{Placed: false},
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: team1},
	))
	assert.False(t, store.SameTeamPlacementForTest(
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: team1},
		store.SwapPlacementForTest{Placed: true, LobbyID: uuid.New(), TeamNumber: team1},
	))
	assert.False(t, store.SameTeamPlacementForTest(
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: nil},
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: team1},
	))
	assert.True(t, store.SameTeamPlacementForTest(
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: team1},
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: team1},
	))
}

func TestSameSubPoolPlacementForTest(t *testing.T) {
	lobbyID := uuid.New()
	otherLobby := uuid.New()
	team1 := teamNumberPtr(1)

	assert.False(t, store.SameSubPoolPlacementForTest(
		store.SwapPlacementForTest{Placed: false},
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: nil},
	))
	assert.True(t, store.SameSubPoolPlacementForTest(
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: nil},
		store.SwapPlacementForTest{Placed: true, LobbyID: otherLobby, TeamNumber: nil},
	))
	assert.False(t, store.SameSubPoolPlacementForTest(
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: team1},
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: nil},
	))
	assert.True(t, store.SameSubPoolPlacementForTest(
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: nil},
		store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: nil},
	))
}

func TestResolveSwapDestinationForTest(t *testing.T) {
	lobbyID := uuid.New()
	team1 := teamNumberPtr(1)
	unplacedSlot := store.SwapPlacementForTest{Placed: false}
	rosterSlot := store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: team1}
	subSlot := store.SwapPlacementForTest{Placed: true, LobbyID: lobbyID, TeamNumber: nil}

	// Taking an unplaced slot with no prior roster seat → unplaced.
	assert.False(t, store.ResolveSwapDestinationForTest(unplacedSlot, unplacedSlot, true).Placed)

	// Non-sub cannot take a sub slot → unplaced.
	nonSubToSub := store.ResolveSwapDestinationForTest(rosterSlot, subSlot, false)
	assert.False(t, nonSubToSub.Placed)

	// Taking a roster seat keeps that seat.
	roster := store.ResolveSwapDestinationForTest(unplacedSlot, rosterSlot, false)
	assert.True(t, roster.Placed)
	assert.Equal(t, lobbyID, roster.LobbyID)
	assert.Equal(t, team1, roster.TeamNumber)

	// Roster player who can sub, displaced by unplaced → own lobby sub pool.
	toSub := store.ResolveSwapDestinationForTest(rosterSlot, unplacedSlot, true)
	assert.True(t, toSub.Placed)
	assert.Equal(t, lobbyID, toSub.LobbyID)
	assert.Nil(t, toSub.TeamNumber)

	// Roster player who cannot sub, displaced by unplaced → unplaced.
	assert.False(t, store.ResolveSwapDestinationForTest(rosterSlot, unplacedSlot, false).Placed)
}

func findLobbyTeamPlayers(detail model.EventGroupDetail, eventID uuid.UUID, teamNumber int) []model.LobbyPlayer {
	var players []model.LobbyPlayer
	for _, event := range detail.Events {
		if event.ID != eventID {
			continue
		}
		for _, lobby := range event.Lobbies {
			for _, team := range lobby.Teams {
				if team.TeamNumber == teamNumber {
					players = append(players, team.Players...)
				}
			}
		}
	}
	return players
}

func findEventInDetail(detail model.EventGroupDetail, eventID uuid.UUID) (model.EventGroupEvent, bool) {
	for _, event := range detail.Events {
		if event.ID == eventID {
			return event, true
		}
	}
	return model.EventGroupEvent{}, false
}

func firstTeamOnePlayerInLobby(lobby model.EventLobby) (model.LobbyPlayer, bool) {
	for _, team := range lobby.Teams {
		if team.TeamNumber != 1 || len(team.Players) == 0 {
			continue
		}
		return team.Players[0], true
	}
	return model.LobbyPlayer{}, false
}

func playerInLobbyTeams(lobby model.EventLobby, userID uuid.UUID) bool {
	for _, team := range lobby.Teams {
		for _, player := range team.Players {
			if player.UserID == userID {
				return true
			}
		}
	}
	return false
}

func findLobbySub(detail model.EventGroupDetail, eventID uuid.UUID) (model.LobbyPlayer, bool) {
	for _, event := range detail.Events {
		if event.ID != eventID {
			continue
		}
		for _, lobby := range event.Lobbies {
			if len(lobby.Subs) > 0 {
				return lobby.Subs[0], true
			}
		}
	}
	return model.LobbyPlayer{}, false
}

func findLobbySubAtCount(detail model.EventGroupDetail, eventID uuid.UUID, n int) (model.LobbyPlayer, bool) {
	for _, event := range detail.Events {
		if event.ID != eventID {
			continue
		}
		for _, lobby := range event.Lobbies {
			if len(lobby.Subs) == n {
				return lobby.Subs[0], true
			}
		}
	}
	return model.LobbyPlayer{}, false
}

func TestSwapPlayersForEvent_SameLobbyTeamSwap(t *testing.T) {
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

	require.NoError(t, s.SwapPlayersForEvent(ctx, eventID, host.ID, p1.UserID, p2.UserID, defaultMatchmakingSettings()))

	after, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)

	var team1HasP2, team2HasP1 bool
	for _, event := range after.Events {
		if event.ID != eventID {
			continue
		}
		for _, lobby := range event.Lobbies {
			for _, team := range lobby.Teams {
				for _, player := range team.Players {
					if team.TeamNumber == 1 && player.UserID == p2.UserID {
						team1HasP2 = true
					}
					if team.TeamNumber == 2 && player.UserID == p1.UserID {
						team2HasP1 = true
					}
				}
			}
		}
	}
	assert.True(t, team1HasP2)
	assert.True(t, team2HasP1)
}

func TestSwapPlayersForEvent_RosterNonSubWithSubBecomesUnplaced(t *testing.T) {
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

	subUser := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subUser.ID, games[0].ID, true, false)
	lobbyID := firstLobbyID(t, ctx, tx, eventID)
	insertPlayerForLobby(t, ctx, tx, lobbyID, subUser.ID, nil)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)

	rosterPlayer, _, ok := findLobbyTeamPlayer(detail, eventID, 1)
	require.True(t, ok)

	require.NoError(t, s.SwapPlayersForEvent(ctx, eventID, host.ID, rosterPlayer.UserID, subUser.ID, defaultMatchmakingSettings()))

	after, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)

	var rosterOnTeam bool
	var formerRosterUnplaced bool
	var subOnRoster bool
	for _, event := range after.Events {
		if event.ID != eventID {
			continue
		}
		for _, lobby := range event.Lobbies {
			for _, team := range lobby.Teams {
				for _, player := range team.Players {
					if player.UserID == rosterPlayer.UserID {
						rosterOnTeam = true
					}
					if player.UserID == subUser.ID {
						subOnRoster = true
					}
				}
			}
		}
		for _, unplaced := range event.Unplaced {
			if unplaced.UserID == rosterPlayer.UserID {
				formerRosterUnplaced = true
			}
		}
	}

	assert.False(t, rosterOnTeam)
	assert.True(t, formerRosterUnplaced)
	assert.True(t, subOnRoster)
}

func TestSwapPlayersForEvent_MultiLobbySubMinViolation(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))
	setGroupSubMin(t, ctx, tx, groupID, 1)

	for i := 0; i < 8; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, false)
	}
	for i := 0; i < 3; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, true, false)
	}

	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)

	subPlayer, ok := findLobbySubAtCount(detail, eventID, 1)
	require.True(t, ok)
	rosterPlayer, _, ok := findLobbyTeamPlayer(detail, eventID, 1)
	require.True(t, ok)

	err = s.SwapPlayersForEvent(ctx, eventID, host.ID, rosterPlayer.UserID, subPlayer.UserID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInsufficientSubstitutes)
}

func TestSwapPlayersForEvent_SingleLobbySkipsSubMinViolation(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))
	setGroupSubMin(t, ctx, tx, groupID, 1)

	for i := 0; i < 4; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, false)
	}

	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	subUser := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subUser.ID, games[0].ID, true, false)
	lobbyID := firstLobbyID(t, ctx, tx, eventID)
	insertPlayerForLobby(t, ctx, tx, lobbyID, subUser.ID, nil)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)

	rosterPlayer, _, ok := findLobbyTeamPlayer(detail, eventID, 1)
	require.True(t, ok)

	require.NoError(t, s.SwapPlayersForEvent(ctx, eventID, host.ID, rosterPlayer.UserID, subUser.ID, defaultMatchmakingSettings()))
}

func TestSwapPlayersForEvent_Forbidden(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	other := createTestUser(t, ctx, s)
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

	err = s.SwapPlayersForEvent(ctx, eventID, other.ID, p1.UserID, p2.UserID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestSwapPlayersForEvent_EventNotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)

	err := s.SwapPlayersForEvent(ctx, uuid.New(), host.ID, uuid.New(), uuid.New(), defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventNotFound)
}

func TestSwapPlayersForEvent_TeamsNotCreated(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))
	p1 := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, p1.ID, games[0].ID, false, false)

	err = s.SwapPlayersForEvent(ctx, eventID, host.ID, p1.ID, uuid.New(), defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrTeamsNotCreated)
}

func TestSwapPlayersForEvent_SameUser(t *testing.T) {
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

	err = s.SwapPlayersForEvent(ctx, eventID, host.ID, p1.UserID, p1.UserID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
	var swapErr *store.SwapValidationError
	require.ErrorAs(t, err, &swapErr)
	assert.Equal(t, "Cannot swap a player with themselves", swapErr.Message)
}

func TestSwapPlayersForEvent_BothUnplaced(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	for i := 0; i < 6; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, false)
	}
	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	require.Len(t, detail.Events[0].Unplaced, 2)

	u1 := detail.Events[0].Unplaced[0].UserID
	u2 := detail.Events[0].Unplaced[1].UserID

	err = s.SwapPlayersForEvent(ctx, eventID, host.ID, u1, u2, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
	var swapErr *store.SwapValidationError
	require.ErrorAs(t, err, &swapErr)
	assert.Equal(t, "Both players are unplaced", swapErr.Message)
}

func TestSwapPlayersForEvent_SameSubPool(t *testing.T) {
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

	lobbyID := firstLobbyID(t, ctx, tx, eventID)
	subA := createTestUser(t, ctx, s)
	subB := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subA.ID, games[0].ID, true, false)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subB.ID, games[0].ID, true, false)
	insertPlayerForLobby(t, ctx, tx, lobbyID, subA.ID, nil)
	insertPlayerForLobby(t, ctx, tx, lobbyID, subB.ID, nil)

	err = s.SwapPlayersForEvent(ctx, eventID, host.ID, subA.ID, subB.ID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
	var swapErr *store.SwapValidationError
	require.ErrorAs(t, err, &swapErr)
	assert.Equal(t, "Cannot swap players in the same sub pool", swapErr.Message)
}

func TestSwapPlayersForEvent_CrossLobbySameSubPool(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	for i := 0; i < 8; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, false)
	}
	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	eventBefore, ok := findEventInDetail(detail, eventID)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(eventBefore.Lobbies), 2)

	subA := createTestUser(t, ctx, s)
	subB := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subA.ID, games[0].ID, true, false)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subB.ID, games[0].ID, true, false)
	insertPlayerForLobby(t, ctx, tx, eventBefore.Lobbies[0].ID, subA.ID, nil)
	insertPlayerForLobby(t, ctx, tx, eventBefore.Lobbies[1].ID, subB.ID, nil)

	err = s.SwapPlayersForEvent(ctx, eventID, host.ID, subA.ID, subB.ID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
	var swapErr *store.SwapValidationError
	require.ErrorAs(t, err, &swapErr)
	assert.Equal(t, "Cannot swap players in the same sub pool", swapErr.Message)
}

func TestSwapPlayersForEvent_SameTeam(t *testing.T) {
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
	team1Players := findLobbyTeamPlayers(detail, eventID, 1)
	require.GreaterOrEqual(t, len(team1Players), 2)

	err = s.SwapPlayersForEvent(ctx, eventID, host.ID, team1Players[0].UserID, team1Players[1].UserID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
}

func TestSwapPlayersForEvent_NotRegistered(t *testing.T) {
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

	err = s.SwapPlayersForEvent(ctx, eventID, host.ID, p1.UserID, uuid.New(), defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
}

func TestSwapPlayersForEvent_RosterWithUnplaced(t *testing.T) {
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
	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	require.Len(t, detail.Events[0].Unplaced, 1)
	unplacedID := detail.Events[0].Unplaced[0].UserID
	rosterPlayer, _, ok := findLobbyTeamPlayer(detail, eventID, 1)
	require.True(t, ok)

	require.NoError(t, s.SwapPlayersForEvent(ctx, eventID, host.ID, rosterPlayer.UserID, unplacedID, defaultMatchmakingSettings()))

	after, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)

	var rosterNowUnplaced bool
	var unplacedNowRoster bool
	for _, event := range after.Events {
		if event.ID != eventID {
			continue
		}
		for _, lobby := range event.Lobbies {
			for _, team := range lobby.Teams {
				for _, player := range team.Players {
					if player.UserID == unplacedID {
						unplacedNowRoster = true
					}
				}
			}
		}
		for _, unplaced := range event.Unplaced {
			if unplaced.UserID == rosterPlayer.UserID {
				rosterNowUnplaced = true
			}
		}
	}
	assert.True(t, rosterNowUnplaced)
	assert.True(t, unplacedNowRoster)
}

func TestSwapPlayersForEvent_RosterCanSubWithUnplacedBecomesSub(t *testing.T) {
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
	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	require.Len(t, detail.Events[0].Unplaced, 1)
	unplacedID := detail.Events[0].Unplaced[0].UserID
	rosterPlayer, lobbyID, ok := findLobbyTeamPlayer(detail, eventID, 1)
	require.True(t, ok)

	_, err = tx.Exec(ctx, `
		UPDATE registrations SET can_substitute = true WHERE event_id = $1 AND user_id = $2
	`, eventID, rosterPlayer.UserID)
	require.NoError(t, err)

	require.NoError(t, s.SwapPlayersForEvent(ctx, eventID, host.ID, rosterPlayer.UserID, unplacedID, defaultMatchmakingSettings()))

	after, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)

	var rosterNowSub bool
	var unplacedNowRoster bool
	for _, event := range after.Events {
		if event.ID != eventID {
			continue
		}
		for _, lobby := range event.Lobbies {
			for _, team := range lobby.Teams {
				for _, player := range team.Players {
					if player.UserID == unplacedID {
						unplacedNowRoster = true
					}
				}
			}
			if lobby.ID == lobbyID {
				for _, sub := range lobby.Subs {
					if sub.UserID == rosterPlayer.UserID {
						rosterNowSub = true
					}
				}
			}
		}
		for _, stillUnplaced := range event.Unplaced {
			assert.NotEqual(t, rosterPlayer.UserID, stillUnplaced.UserID)
		}
	}
	assert.True(t, rosterNowSub)
	assert.True(t, unplacedNowRoster)
}

func TestSwapPlayersForEvent_CrossLobby(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	for i := 0; i < 8; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, false)
	}
	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	eventBefore, ok := findEventInDetail(detail, eventID)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(eventBefore.Lobbies), 2)

	lobbyA := eventBefore.Lobbies[0]
	lobbyB := eventBefore.Lobbies[1]
	lobby1Player, ok := firstTeamOnePlayerInLobby(lobbyA)
	require.True(t, ok)
	lobby2Player, ok := firstTeamOnePlayerInLobby(lobbyB)
	require.True(t, ok)

	require.NoError(t, s.SwapPlayersForEvent(ctx, eventID, host.ID, lobby1Player.UserID, lobby2Player.UserID, defaultMatchmakingSettings()))

	after, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	eventAfter, ok := findEventInDetail(after, eventID)
	require.True(t, ok)

	var lobbyAOut, lobbyBOut *model.EventLobby
	for i := range eventAfter.Lobbies {
		switch eventAfter.Lobbies[i].ID {
		case lobbyA.ID:
			lobbyAOut = &eventAfter.Lobbies[i]
		case lobbyB.ID:
			lobbyBOut = &eventAfter.Lobbies[i]
		}
	}
	require.NotNil(t, lobbyAOut)
	require.NotNil(t, lobbyBOut)
	assert.True(t, playerInLobbyTeams(*lobbyAOut, lobby2Player.UserID))
	assert.True(t, playerInLobbyTeams(*lobbyBOut, lobby1Player.UserID))
}

func TestSwapPlayersForEvent_SwapRosterWhileSubRemains(t *testing.T) {
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

	subVolunteer := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subVolunteer.ID, games[0].ID, true, true)
	lobbyID := firstLobbyID(t, ctx, tx, eventID)
	insertPlayerForLobby(t, ctx, tx, lobbyID, subVolunteer.ID, nil)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	team1Players := findLobbyTeamPlayers(detail, eventID, 1)
	team2Players := findLobbyTeamPlayers(detail, eventID, 2)
	require.NotEmpty(t, team1Players)
	require.NotEmpty(t, team2Players)

	require.NoError(t, s.SwapPlayersForEvent(ctx, eventID, host.ID, team1Players[0].UserID, team2Players[0].UserID, defaultMatchmakingSettings()))

	var lobbyHost uuid.UUID
	var fairnessWarning bool
	err = tx.QueryRow(ctx, `SELECT host, fairness_warning FROM lobbies WHERE id = $1`, lobbyID).Scan(&lobbyHost, &fairnessWarning)
	require.NoError(t, err)
	assert.Equal(t, subVolunteer.ID, lobbyHost)

	var subCount int
	err = tx.QueryRow(ctx, `SELECT count(*) FROM players WHERE lobby_id = $1 AND team_number IS NULL`, lobbyID).Scan(&subCount)
	require.NoError(t, err)
	assert.Equal(t, 1, subCount)
}

func TestRecomputeLobbyAfterSwapForTest_IncludesSubsAndVolunteerHost(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	lobbyID := insertLobbyForEvent(t, ctx, tx, eventID, &host.ID)
	team1 := teamNumberPtr(1)
	team2 := teamNumberPtr(2)
	for i := 0; i < 4; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, false)
		tn := team1
		if i >= 2 {
			tn = team2
		}
		insertPlayerForLobby(t, ctx, tx, lobbyID, u.ID, tn)
	}

	subVolunteer := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subVolunteer.ID, games[0].ID, true, true)
	insertPlayerForLobby(t, ctx, tx, lobbyID, subVolunteer.ID, nil)

	require.NoError(t, store.RecomputeLobbyAfterSwapForTest(s, ctx, lobbyID, defaultMatchmakingSettings(), modeID))

	var lobbyHost uuid.UUID
	var fairnessWarning bool
	err = tx.QueryRow(ctx, `SELECT host, fairness_warning FROM lobbies WHERE id = $1`, lobbyID).Scan(&lobbyHost, &fairnessWarning)
	require.NoError(t, err)
	assert.Equal(t, subVolunteer.ID, lobbyHost)
	_ = fairnessWarning
}

func TestRecomputeLobbyAfterSwapForTest_InvalidGameMode(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))
	lobbyID := insertLobbyForEvent(t, ctx, tx, eventID, &host.ID)

	err = store.RecomputeLobbyAfterSwapForTest(s, ctx, lobbyID, defaultMatchmakingSettings(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get game mode")
}
