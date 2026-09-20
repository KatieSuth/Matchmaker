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

func TestPickLobbyHost_IgnoresSubVolunteer(t *testing.T) {
	teamVolunteer := uuid.New()
	teamFallback := uuid.New()
	subVolunteer := uuid.New()
	now := time.Now()
	lobby := matchmaking.LobbyPlan{
		Roster: []matchmaking.Player{
			{UserID: teamFallback, CanLobbyHost: false, CreatedAt: now},
			{UserID: teamVolunteer, CanLobbyHost: true, CreatedAt: now.Add(time.Minute)},
		},
		Subs: []matchmaking.Player{
			{UserID: subVolunteer, CanLobbyHost: true, CreatedAt: now.Add(-time.Hour)},
		},
	}
	host := matchmaking.PickLobbyHost(lobby)
	require.NotNil(t, host)
	assert.Equal(t, teamVolunteer, *host)
}

func TestPickLobbyHost_FallsBackToTeamPlayerWhenOnlySubVolunteered(t *testing.T) {
	firstTeam := uuid.New()
	secondTeam := uuid.New()
	subVolunteer := uuid.New()
	now := time.Now()
	lobby := matchmaking.LobbyPlan{
		Roster: []matchmaking.Player{
			{UserID: firstTeam, CanLobbyHost: false, CreatedAt: now},
			{UserID: secondTeam, CanLobbyHost: false, CreatedAt: now.Add(time.Minute)},
		},
		Subs: []matchmaking.Player{
			{UserID: subVolunteer, CanLobbyHost: true, CreatedAt: now.Add(-time.Hour)},
		},
	}
	host := matchmaking.PickLobbyHost(lobby)
	require.NotNil(t, host)
	assert.Equal(t, firstTeam, *host)
}

func TestPickLobbyHost_EmptyRosterReturnsNil(t *testing.T) {
	lobby := matchmaking.LobbyPlan{
		Subs: []matchmaking.Player{
			{UserID: uuid.New(), CanLobbyHost: true, CreatedAt: time.Now()},
		},
	}
	assert.Nil(t, matchmaking.PickLobbyHost(lobby))
}
