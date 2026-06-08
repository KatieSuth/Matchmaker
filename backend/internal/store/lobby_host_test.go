package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLobbyHostValidationError_ErrorAndUnwrap(t *testing.T) {
	err := &store.LobbyHostValidationError{Message: "Player is already the lobby host"}
	assert.Equal(t, "Player is already the lobby host", err.Error())
	assert.ErrorIs(t, err, store.ErrInvalidLobbyHostChange)
}

func TestSetLobbyHostForEvent_Success(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	for i := 0; i < 4; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, i == 0)
	}

	require.NoError(t, s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings()))

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)

	target, lobbyID, ok := findLobbyTeamPlayer(detail, eventID, 2)
	require.True(t, ok)

	require.NoError(t, s.SetLobbyHostForEvent(ctx, eventID, host.ID, target.UserID))

	var lobbyHost uuid.UUID
	err = tx.QueryRow(ctx, `SELECT host FROM lobbies WHERE id = $1`, lobbyID).Scan(&lobbyHost)
	require.NoError(t, err)
	assert.Equal(t, target.UserID, lobbyHost)
}

func TestSetLobbyHostForEvent_Forbidden(t *testing.T) {
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

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	target, _, ok := findLobbyTeamPlayer(detail, eventID, 1)
	require.True(t, ok)

	err = s.SetLobbyHostForEvent(ctx, eventID, other.ID, target.UserID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestSetLobbyHostForEvent_AlreadyHost(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	for i := 0; i < 4; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, i == 0)
	}

	require.NoError(t, s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings()))

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)

	var currentHost uuid.UUID
	for _, event := range detail.Events {
		if event.ID != eventID {
			continue
		}
		for _, lobby := range event.Lobbies {
			require.NotNil(t, lobby.HostID)
			currentHost = *lobby.HostID
		}
	}

	err = s.SetLobbyHostForEvent(ctx, eventID, host.ID, currentHost)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidLobbyHostChange)
	var hostErr *store.LobbyHostValidationError
	require.ErrorAs(t, err, &hostErr)
	assert.Equal(t, "Player is already the lobby host", hostErr.Message)
}

func TestSetLobbyHostForEvent_SubPlayerRejected(t *testing.T) {
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

	subUser := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, subUser.ID, games[0].ID, true, true)

	lobbyID := firstLobbyID(t, ctx, tx, eventID)
	insertPlayerForLobby(t, ctx, tx, lobbyID, subUser.ID, nil)

	err = s.SetLobbyHostForEvent(ctx, eventID, host.ID, subUser.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidLobbyHostChange)
	var hostErr *store.LobbyHostValidationError
	require.ErrorAs(t, err, &hostErr)
	assert.Equal(t, "Only team players can be made lobby host", hostErr.Message)
}

func TestSetLobbyHostForEvent_UnplacedPlayerRejected(t *testing.T) {
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

	unplaced := createTestUser(t, ctx, s)
	registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, unplaced.ID, games[0].ID, false, true)

	err = s.SetLobbyHostForEvent(ctx, eventID, host.ID, unplaced.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidLobbyHostChange)
	var hostErr *store.LobbyHostValidationError
	require.ErrorAs(t, err, &hostErr)
	assert.Equal(t, "Player is not assigned to a team", hostErr.Message)
}

func TestSetLobbyHostForEvent_TeamsNotCreated(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	err = s.SetLobbyHostForEvent(ctx, eventID, host.ID, uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrTeamsNotCreated)
}

func TestSetLobbyHostForEvent_EventNotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)

	err := s.SetLobbyHostForEvent(ctx, uuid.New(), host.ID, uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventNotFound)
}

func TestSetLobbyHostForEvent_ContextCancelled(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.SetLobbyHostForEvent(ctx, uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
}

func TestSetLobbyHostForEvent_NotRegistered(t *testing.T) {
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

	err = s.SetLobbyHostForEvent(ctx, eventID, host.ID, uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidLobbyHostChange)
	var hostErr *store.LobbyHostValidationError
	require.ErrorAs(t, err, &hostErr)
	assert.Equal(t, "One or both players are not registered for this game", hostErr.Message)
}

func TestSetLobbyHostForEvent_LobbyHostUnset(t *testing.T) {
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
	target, lobbyID, ok := findLobbyTeamPlayer(detail, eventID, 2)
	require.True(t, ok)

	_, err = tx.Exec(ctx, `UPDATE lobbies SET host = NULL WHERE id = $1`, lobbyID)
	require.NoError(t, err)

	err = s.SetLobbyHostForEvent(ctx, eventID, host.ID, target.UserID)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidLobbyHostChange)
	var hostErr *store.LobbyHostValidationError
	require.ErrorAs(t, err, &hostErr)
	assert.Equal(t, "Lobby not found for this player", hostErr.Message)
}

func TestSetLobbyHostForEvent_MultiLobbySuccess(t *testing.T) {
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

	require.NoError(t, s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings()))

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(detail.Events[0].Lobbies), 2)

	target, lobbyID, ok := findNonHostTeamPlayerInLobbyIndex(detail, eventID, 1)
	require.True(t, ok)

	require.NoError(t, s.SetLobbyHostForEvent(ctx, eventID, host.ID, target.UserID))

	var lobbyHost uuid.UUID
	err = tx.QueryRow(ctx, `SELECT host FROM lobbies WHERE id = $1`, lobbyID).Scan(&lobbyHost)
	require.NoError(t, err)
	assert.Equal(t, target.UserID, lobbyHost)
}

func findNonHostTeamPlayerInLobbyIndex(
	detail model.EventGroupDetail,
	eventID uuid.UUID,
	lobbyIndex int,
) (model.LobbyPlayer, uuid.UUID, bool) {
	for _, event := range detail.Events {
		if event.ID != eventID {
			continue
		}
		if lobbyIndex >= len(event.Lobbies) {
			return model.LobbyPlayer{}, uuid.Nil, false
		}
		lobby := event.Lobbies[lobbyIndex]
		for _, team := range lobby.Teams {
			for _, player := range team.Players {
				if lobby.HostID != nil && player.UserID == *lobby.HostID {
					continue
				}
				return player, lobby.ID, true
			}
		}
	}
	return model.LobbyPlayer{}, uuid.Nil, false
}
