package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwapPlayersForEvent_SubWithUnplacedRejected(t *testing.T) {
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

	lobbyID := firstLobbyID(t, ctx, tx, eventID)
	subUser := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subUser.ID, games[0].ID, true, false)
	insertPlayerForLobby(t, ctx, tx, lobbyID, subUser.ID, nil)

	err = s.SwapPlayersForEvent(ctx, eventID, host.ID, subUser.ID, unplacedID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
	var swapErr *store.SwapValidationError
	require.ErrorAs(t, err, &swapErr)
	assert.Equal(t, "Cannot swap between substitutes and unplaced players", swapErr.Message)
}

func TestMoveSubToUnplacedForEvent_Success(t *testing.T) {
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

	require.NoError(t, s.MoveSubToUnplacedForEvent(ctx, eventID, host.ID, subUser.ID, defaultMatchmakingSettings()))

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	event, ok := findEventInDetail(detail, eventID)
	require.True(t, ok)
	assert.True(t, playerInLobbyTeams(event.Lobbies[0], subUser.ID) == false)
	foundUnplaced := false
	for _, u := range event.Unplaced {
		if u.UserID == subUser.ID {
			foundUnplaced = true
		}
	}
	assert.True(t, foundUnplaced)
}

func TestMoveUnplacedToSubsForEvent_Success(t *testing.T) {
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
	event, ok := findEventInDetail(detail, eventID)
	require.True(t, ok)
	require.Len(t, event.Unplaced, 1)
	unplacedID := event.Unplaced[0].UserID
	_, err = tx.Exec(ctx, `
		UPDATE registrations SET can_substitute = true
		WHERE event_id = $1 AND user_id = $2`, eventID, unplacedID)
	require.NoError(t, err)

	lobbyID := event.Lobbies[0].ID
	require.NoError(t, s.MoveUnplacedToSubsForEvent(ctx, eventID, host.ID, unplacedID, lobbyID, defaultMatchmakingSettings()))

	after, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	afterEvent, ok := findEventInDetail(after, eventID)
	require.True(t, ok)
	assert.Empty(t, afterEvent.Unplaced)
	require.Len(t, afterEvent.Lobbies[0].Subs, 1)
	assert.Equal(t, unplacedID, afterEvent.Lobbies[0].Subs[0].UserID)
}

func TestMoveUnplacedToSubsForEvent_NotSubstituteEligible(t *testing.T) {
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
	lobbyID := detail.Events[0].Lobbies[0].ID

	err = s.MoveUnplacedToSubsForEvent(ctx, eventID, host.ID, detail.Events[0].Unplaced[0].UserID, lobbyID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
}

func TestMoveSubToUnplacedForEvent_Forbidden(t *testing.T) {
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
	require.NoError(t, s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings()))

	lobbyID := firstLobbyID(t, ctx, tx, eventID)
	subUser := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subUser.ID, games[0].ID, true, false)
	insertPlayerForLobby(t, ctx, tx, lobbyID, subUser.ID, nil)

	err = s.MoveSubToUnplacedForEvent(ctx, eventID, other.ID, subUser.ID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestMoveSubToUnplacedForEvent_EventNotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)

	err := s.MoveSubToUnplacedForEvent(ctx, uuid.New(), host.ID, uuid.New(), defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventNotFound)
}

func TestMoveSubToUnplacedForEvent_TeamsNotCreated(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))
	subUser := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subUser.ID, games[0].ID, true, false)

	err = s.MoveSubToUnplacedForEvent(ctx, eventID, host.ID, subUser.ID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrTeamsNotCreated)
}

func TestMoveSubToUnplacedForEvent_NotRegistered(t *testing.T) {
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

	err = s.MoveSubToUnplacedForEvent(ctx, eventID, host.ID, uuid.New(), defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
}

func TestMoveSubToUnplacedForEvent_NotSubstitute(t *testing.T) {
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

	rosterPlayer, _, ok := findLobbyTeamPlayer(detail, eventID, 1)
	require.True(t, ok)
	require.Len(t, detail.Events[0].Unplaced, 1)
	unplacedID := detail.Events[0].Unplaced[0].UserID

	err = s.MoveSubToUnplacedForEvent(ctx, eventID, host.ID, rosterPlayer.UserID, defaultMatchmakingSettings())
	require.Error(t, err)
	var swapErr *store.SwapValidationError
	require.ErrorAs(t, err, &swapErr)
	assert.Equal(t, "Player is not a substitute", swapErr.Message)

	err = s.MoveSubToUnplacedForEvent(ctx, eventID, host.ID, unplacedID, defaultMatchmakingSettings())
	require.Error(t, err)
	require.ErrorAs(t, err, &swapErr)
	assert.Equal(t, "Player is not a substitute", swapErr.Message)
}

func TestMoveSubToUnplacedForEvent_MultiLobbySubMinViolation(t *testing.T) {
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
	for i := 0; i < 2; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, true, false)
	}
	require.NoError(t, s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings()))

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	subPlayer, ok := findLobbySub(detail, eventID)
	require.True(t, ok)

	err = s.MoveSubToUnplacedForEvent(ctx, eventID, host.ID, subPlayer.UserID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInsufficientSubstitutes)
}

func TestMoveUnplacedToSubsForEvent_Forbidden(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	other := createTestUser(t, ctx, s)
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
	lobbyID := detail.Events[0].Lobbies[0].ID

	err = s.MoveUnplacedToSubsForEvent(ctx, eventID, other.ID, unplacedID, lobbyID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestMoveUnplacedToSubsForEvent_EventNotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)

	err := s.MoveUnplacedToSubsForEvent(ctx, uuid.New(), host.ID, uuid.New(), uuid.New(), defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventNotFound)
}

func TestMoveUnplacedToSubsForEvent_TeamsNotCreated(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))
	unplaced := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, unplaced.ID, games[0].ID, true, false)

	err = s.MoveUnplacedToSubsForEvent(ctx, eventID, host.ID, unplaced.ID, uuid.New(), defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrTeamsNotCreated)
}

func TestMoveUnplacedToSubsForEvent_NotRegistered(t *testing.T) {
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
	lobbyID := firstLobbyID(t, ctx, tx, eventID)

	err = s.MoveUnplacedToSubsForEvent(ctx, eventID, host.ID, uuid.New(), lobbyID, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidPlayerSwap)
}

func TestMoveUnplacedToSubsForEvent_AlreadyPlaced(t *testing.T) {
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

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	rosterPlayer, lobbyID, ok := findLobbyTeamPlayer(detail, eventID, 1)
	require.True(t, ok)
	_, err = tx.Exec(ctx, `
		UPDATE registrations SET can_substitute = true
		WHERE event_id = $1 AND user_id = $2`, eventID, rosterPlayer.UserID)
	require.NoError(t, err)

	err = s.MoveUnplacedToSubsForEvent(ctx, eventID, host.ID, rosterPlayer.UserID, lobbyID, defaultMatchmakingSettings())
	require.Error(t, err)
	var swapErr *store.SwapValidationError
	require.ErrorAs(t, err, &swapErr)
	assert.Equal(t, "Player is already placed", swapErr.Message)
}

func TestMoveUnplacedToSubsForEvent_LobbyNotFound(t *testing.T) {
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

	err = s.MoveUnplacedToSubsForEvent(ctx, eventID, host.ID, unplacedID, uuid.New(), defaultMatchmakingSettings())
	require.Error(t, err)
	var swapErr *store.SwapValidationError
	require.ErrorAs(t, err, &swapErr)
	assert.Equal(t, "Lobby not found for this game", swapErr.Message)
}
