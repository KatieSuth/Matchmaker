package store_test

import (
	"context"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamCreationError_ErrorReturnsMessage(t *testing.T) {
	err := &store.TeamCreationError{
		Sentinel: store.ErrInsufficientPlayers,
		Message:  "Game 1 (2v2 Test) needs at least 4 players but only 1 are registered",
	}
	assert.Equal(t, "Game 1 (2v2 Test) needs at least 4 players but only 1 are registered", err.Error())
	assert.ErrorIs(t, err, store.ErrInsufficientPlayers)
}

func TestIntPtrToInt32PtrForTest_NilAndValue(t *testing.T) {
	assert.Nil(t, store.IntPtrToInt32PtrForTest(nil))
	v := 2
	out := store.IntPtrToInt32PtrForTest(&v)
	require.NotNil(t, out)
	assert.Equal(t, int32(2), *out)
}

func TestPersistTeamPlansForTest_CreateLobbyError(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	team := 1
	hostID := uuid.New()
	err := store.PersistTeamPlansForTest(s, ctx, []matchmaking.GamePlan{{
		EventID: uuid.New(),
		Lobbies: []matchmaking.LobbyPlan{{
			HostID: &hostID,
			Roster: []matchmaking.Player{{
				UserID:     uuid.New(),
				TeamNumber: &team,
			}},
		}},
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create lobby")
}
