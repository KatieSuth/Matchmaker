package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func playerAtRank(rank float64, canSub bool, gameCount int, at time.Time) matchmaking.Player {
	return matchmaking.Player{
		UserID:              uuid.New(),
		AvgRank:             rank,
		CanSubstitute:       canSub,
		RegisteredGameCount: gameCount,
		CreatedAt:           at,
	}
}

func TestSelectBalancedRosterPool_CyclesHighMidLow(t *testing.T) {
	now := time.Now()
	high := playerAtRank(24, false, 1, now)
	mid := playerAtRank(10, false, 1, now.Add(time.Minute))
	low := playerAtRank(4, true, 1, now.Add(2*time.Minute))
	players := []matchmaking.Player{
		high,
		playerAtRank(16, false, 1, now.Add(3*time.Minute)),
		mid,
		playerAtRank(6, false, 1, now.Add(4*time.Minute)),
		low,
	}

	lobbies := matchmaking.AssignBalanced(players, 1, 3)
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 3)

	rosteredIDs := make(map[uuid.UUID]bool)
	for _, p := range lobbies[0].Roster {
		rosteredIDs[p.UserID] = true
	}
	assert.True(t, rosteredIDs[high.UserID])
	assert.True(t, rosteredIDs[mid.UserID], "balanced roster should include mid-skill players")
	assert.True(t, rosteredIDs[low.UserID])
}

func TestSelectBalancedRosterPool_SpansHighAndLowSkill(t *testing.T) {
	now := time.Now()
	lowSub := playerAtRank(4, true, 1, now)
	players := []matchmaking.Player{
		playerAtRank(24, false, 1, now.Add(time.Minute)),
		playerAtRank(23, false, 1, now.Add(2*time.Minute)),
		playerAtRank(22, false, 1, now.Add(3*time.Minute)),
		playerAtRank(21, false, 1, now.Add(4*time.Minute)),
		playerAtRank(20, false, 1, now.Add(5*time.Minute)),
		lowSub,
	}

	lobbies := matchmaking.AssignBalanced(players, 1, 4)
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 4)

	rosteredIDs := make(map[uuid.UUID]bool)
	for _, p := range lobbies[0].Roster {
		rosteredIDs[p.UserID] = true
	}
	assert.True(t, rosteredIDs[lowSub.UserID], "balanced roster should include low-skill players")
}

func TestSelectBalancedRosterPool_SubstituteEligibilityDoesNotAffectSingleLobbyRoster(t *testing.T) {
	now := time.Now()
	highSub := playerAtRank(20, true, 1, now)
	lowNonSub := playerAtRank(4, false, 1, now.Add(time.Minute))
	players := []matchmaking.Player{highSub, lowNonSub}

	lobbies := matchmaking.AssignBalanced(players, 1, 1)
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 1)
	assert.Equal(t, highSub.UserID, lobbies[0].Roster[0].UserID)
}

func TestSelectRankedRosterPool_KeepsLowMajority(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{
		playerAtRank(3, false, 1, now),
		playerAtRank(4, false, 1, now.Add(time.Minute)),
		playerAtRank(5, false, 1, now.Add(2*time.Minute)),
		playerAtRank(6, false, 1, now.Add(3*time.Minute)),
		playerAtRank(20, false, 1, now.Add(4*time.Minute)),
		playerAtRank(21, false, 1, now.Add(5*time.Minute)),
	}

	lobbies := matchmaking.AssignRanked(players, 1, 4)
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 4)

	for _, p := range lobbies[0].Roster {
		assert.LessOrEqual(t, p.AvgRank, 6.0)
	}
}

func TestSelectRankedRosterPool_KeepsHighMajority(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{
		playerAtRank(3, false, 1, now),
		playerAtRank(4, false, 1, now.Add(time.Minute)),
		playerAtRank(20, false, 1, now.Add(2*time.Minute)),
		playerAtRank(21, false, 1, now.Add(3*time.Minute)),
		playerAtRank(22, false, 1, now.Add(4*time.Minute)),
		playerAtRank(23, false, 1, now.Add(5*time.Minute)),
	}

	lobbies := matchmaking.AssignRanked(players, 1, 4)
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 4)

	for _, p := range lobbies[0].Roster {
		assert.GreaterOrEqual(t, p.AvgRank, 20.0)
	}
}
