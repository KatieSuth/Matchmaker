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

func TestAssignBalanced_SubstituteEligibilityDoesNotAffectSelection(t *testing.T) {
	now := time.Now()
	highSub := playerAtRank(21, true, 1, now)
	highNon := playerAtRank(22, false, 1, now.Add(time.Minute))
	lowSub := playerAtRank(3, true, 1, now.Add(2*time.Minute))
	lowNon := playerAtRank(4, false, 1, now.Add(3*time.Minute))
	players := []matchmaking.Player{highSub, highNon, lowSub, lowNon}

	lobbies := matchmaking.AssignBalanced(players, matchmaking.Config{
		TeamSize:   2,
		LobbyCount: 1,
		TierCount:  25,
		Slots:      4,
	})
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 4)
	ids := map[uuid.UUID]bool{}
	for _, p := range lobbies[0].Roster {
		ids[p.UserID] = true
	}
	assert.True(t, ids[highSub.UserID])
	assert.True(t, ids[highNon.UserID])
	assert.True(t, ids[lowSub.UserID])
	assert.True(t, ids[lowNon.UserID])
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
