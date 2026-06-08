package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPickLobbyHost_PrefersVolunteer(t *testing.T) {
	volunteer := uuid.New()
	fallback := uuid.New()
	now := time.Now()
	lobby := matchmaking.LobbyPlan{
		Roster: []matchmaking.Player{
			{UserID: fallback, CanLobbyHost: false, CreatedAt: now},
			{UserID: volunteer, CanLobbyHost: true, CreatedAt: now.Add(time.Minute)},
		},
	}
	host := matchmaking.PickLobbyHost(lobby)
	require.NotNil(t, host)
	assert.Equal(t, volunteer, *host)
}

func TestPickLobbyHost_FallsBackToFirstMember(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	now := time.Now()
	lobby := matchmaking.LobbyPlan{
		Roster: []matchmaking.Player{
			{UserID: first, CreatedAt: now},
			{UserID: second, CreatedAt: now.Add(time.Minute)},
		},
	}
	host := matchmaking.PickLobbyHost(lobby)
	require.NotNil(t, host)
	assert.Equal(t, first, *host)
}
