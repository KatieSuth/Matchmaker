package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/lobbyjoin"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrString(s string) *string { return &s }

func setupLobbyJoinFixture(t *testing.T) (*store.PostgresStore, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)
	_, err = tx.Exec(ctx, `UPDATE games SET join_link_base = 'https://gg.riotgames.com' WHERE id = $1`, games[0].ID)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	for i := 0; i < 4; i++ {
		u := createTestUser(t, ctx, s)
		registerPlayerForEventWithProfile(t, ctx, tx, s, eventID, u.ID, games[0].ID, false, i == 0)
	}
	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	require.NotNil(t, detail.JoinLinkBase)
	assert.Equal(t, "https://gg.riotgames.com", *detail.JoinLinkBase)

	var lobbyID uuid.UUID
	var lobbyHostID uuid.UUID
	for _, event := range detail.Events {
		if event.ID != eventID {
			continue
		}
		require.NotEmpty(t, event.Lobbies)
		lobbyID = event.Lobbies[0].ID
		require.NotNil(t, event.Lobbies[0].HostID)
		lobbyHostID = *event.Lobbies[0].HostID
	}
	return s, host.ID, lobbyHostID, lobbyID, groupID
}

func TestUpdateLobbyJoinCode_OwnerSetsRiotURL(t *testing.T) {
	s, ownerID, _, lobbyID, groupID := setupLobbyJoinFixture(t)
	ctx := context.Background()

	raw := "https://gg.riotgames.com/LOL?joinCode=FWXy-7C7m-3KQN"
	require.NoError(t, s.UpdateLobbyJoinCode(ctx, lobbyID, ownerID, &raw))

	detail, err := s.GetEventGroupDetail(ctx, groupID, ownerID)
	require.NoError(t, err)
	var found *string
	for _, event := range detail.Events {
		for _, lobby := range event.Lobbies {
			if lobby.ID == lobbyID {
				found = lobby.JoinCode
			}
		}
	}
	require.NotNil(t, found)
	assert.Equal(t, "/LOL?joinCode=FWXy-7C7m-3KQN", *found)
}

func TestUpdateLobbyJoinCode_LobbyHostSetsPlainCode(t *testing.T) {
	s, _, lobbyHostID, lobbyID, groupID := setupLobbyJoinFixture(t)
	ctx := context.Background()

	raw := "JHL829"
	require.NoError(t, s.UpdateLobbyJoinCode(ctx, lobbyID, lobbyHostID, &raw))

	detail, err := s.GetEventGroupDetail(ctx, groupID, lobbyHostID)
	require.NoError(t, err)
	var found *string
	for _, event := range detail.Events {
		for _, lobby := range event.Lobbies {
			if lobby.ID == lobbyID {
				found = lobby.JoinCode
			}
		}
	}
	require.NotNil(t, found)
	assert.Equal(t, "JHL829", *found)
}

func TestUpdateLobbyJoinCode_Clear(t *testing.T) {
	s, ownerID, _, lobbyID, _ := setupLobbyJoinFixture(t)
	ctx := context.Background()

	raw := "ABC123"
	require.NoError(t, s.UpdateLobbyJoinCode(ctx, lobbyID, ownerID, &raw))
	empty := ""
	require.NoError(t, s.UpdateLobbyJoinCode(ctx, lobbyID, ownerID, &empty))

	// nil clears as well
	raw2 := "XYZ"
	require.NoError(t, s.UpdateLobbyJoinCode(ctx, lobbyID, ownerID, &raw2))
	require.NoError(t, s.UpdateLobbyJoinCode(ctx, lobbyID, ownerID, nil))
}

func TestUpdateLobbyJoinCode_Forbidden(t *testing.T) {
	s, _, _, lobbyID, _ := setupLobbyJoinFixture(t)
	ctx := context.Background()
	other := createTestUser(t, ctx, s)

	raw := "ABC123"
	err := s.UpdateLobbyJoinCode(ctx, lobbyID, other.ID, &raw)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrForbidden)
}

func TestUpdateLobbyJoinCode_RejectBadLink(t *testing.T) {
	s, ownerID, _, lobbyID, _ := setupLobbyJoinFixture(t)
	ctx := context.Background()

	raw := "gg.badactor.net/somecode"
	err := s.UpdateLobbyJoinCode(ctx, lobbyID, ownerID, &raw)
	require.Error(t, err)
	assert.ErrorIs(t, err, lobbyjoin.ErrInvalidJoinCode)
}

func TestUpdateLobbyJoinCode_NotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	err := s.UpdateLobbyJoinCode(ctx, uuid.New(), uuid.New(), ptrString("ABC"))
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrLobbyNotFound)
}
