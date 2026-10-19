package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lobbyOfTeams seats team1 and team2 ranks in one lobby.
func lobbyOfTeams(now time.Time, team1, team2 []float64) matchmaking.LobbyPlan {
	var roster []matchmaking.Player
	for _, r := range team1 {
		roster = append(roster, seated(r, 1, now))
	}
	for _, r := range team2 {
		roster = append(roster, seated(r, 2, now))
	}
	return matchmaking.LobbyPlan{Roster: roster}
}

// teamAverages returns each lobby's (team 1, team 2) averages.
func teamAverages(lobbies []matchmaking.LobbyPlan) [][2]float64 {
	out := make([][2]float64, len(lobbies))
	for i, l := range lobbies {
		a, b := matchmaking.SplitRosterByTeamNumberForTest(l.Roster)
		var sa, sb float64
		for _, p := range a {
			sa += p.AvgRank
		}
		for _, p := range b {
			sb += p.AvgRank
		}
		out[i] = [2]float64{sa / float64(len(a)), sb / float64(len(b))}
	}
	return out
}

// The reported case: Plat 2 vs Gold 2 in one lobby and Gold 3 vs Plat 2 in the other become
// Plat 2 vs Plat 2 and Gold 2 vs Gold 3.
func TestPairTeamsAcrossLobbies_ClosestAveragesFaceEachOther(t *testing.T) {
	now := time.Now()
	lobbies := []matchmaking.LobbyPlan{
		lobbyOfTeams(now, []float64{4, 21, 19, 6, 21}, []float64{4, 18, 23, 6, 6}), // 14.2 vs 11.4
		lobbyOfTeams(now, []float64{24, 22, 4, 4, 6}, []float64{20, 4, 22, 20, 6}), // 12.0 vs 14.4
	}
	out := matchmaking.PairTeamsAcrossLobbiesForTest(lobbies, repairSettings(), 25)

	avgs := teamAverages(out)
	assert.InDelta(t, 14.4, avgs[0][0], 1e-9)
	assert.InDelta(t, 14.2, avgs[0][1], 1e-9)
	assert.InDelta(t, 12.0, avgs[1][0], 1e-9)
	assert.InDelta(t, 11.4, avgs[1][1], 1e-9)
	for _, l := range out {
		assert.Len(t, l.Roster, 10, "teams move intact")
		assert.False(t, matchmaking.IsLobbyUnfair(l, repairSettings(), 25))
	}
}

func TestPairTeamsAcrossLobbies_KeepsAlreadyBestPairing(t *testing.T) {
	now := time.Now()
	lobbies := []matchmaking.LobbyPlan{
		lobbyOfTeams(now, []float64{14, 14}, []float64{14, 13}),
		lobbyOfTeams(now, []float64{9, 9}, []float64{10, 9}),
	}
	out := matchmaking.PairTeamsAcrossLobbiesForTest(lobbies, repairSettings(), 25)
	assert.Equal(t, lobbies, out)
}

func TestPairTeamsAcrossLobbies_RejectsPairingThatMakesLobbiesUnfair(t *testing.T) {
	now := time.Now()
	// Re-pairing would cut the worst team gap from 11.5 to 5.5, but leave both lobbies unfair.
	lobbies := []matchmaking.LobbyPlan{
		lobbyOfTeams(now, []float64{19, 14}, []float64{14, 14}),
		lobbyOfTeams(now, []float64{19, 21}, []float64{15, 2}),
	}
	out := matchmaking.PairTeamsAcrossLobbiesForTest(lobbies, repairSettings(), 25)
	assert.Equal(t, lobbies, out)
}

func TestPairTeamsAcrossLobbies_LeavesSingleLobbyAndSeatedSubsAlone(t *testing.T) {
	now := time.Now()
	single := []matchmaking.LobbyPlan{lobbyOfTeams(now, []float64{20, 20}, []float64{5, 5})}
	assert.Equal(t, single, matchmaking.PairTeamsAcrossLobbiesForTest(single, repairSettings(), 25))

	withSubs := []matchmaking.LobbyPlan{
		lobbyOfTeams(now, []float64{14, 14}, []float64{9, 9}),
		lobbyOfTeams(now, []float64{9, 10}, []float64{14, 13}),
	}
	withSubs[0].Subs = []matchmaking.Player{player(uuid.New(), 8, true, now)}
	assert.Equal(t, withSubs, matchmaking.PairTeamsAcrossLobbiesForTest(withSubs, repairSettings(), 25))

	oneSided := []matchmaking.LobbyPlan{
		lobbyOfTeams(now, []float64{14, 14}, nil),
		lobbyOfTeams(now, []float64{9, 10}, []float64{14, 13}),
	}
	assert.Equal(t, oneSided, matchmaking.PairTeamsAcrossLobbiesForTest(oneSided, repairSettings(), 25))
}

func TestPlanEvent_RankedDoesNotRepairTeamsAcrossLobbies(t *testing.T) {
	now := time.Now()
	var players []matchmaking.Player
	for i, r := range []float64{2, 3, 4, 5, 20, 21, 22, 23} {
		players = append(players, matchmaking.Player{UserID: uuid.New(), AvgRank: r, CreatedAt: now.Add(time.Duration(i) * time.Second)})
	}
	plan, err := matchmaking.PlanEvent(players, matchmaking.Config{
		EventID: uuid.New(), TeamSize: 2, Slots: 4, TierCount: 25, SortLogic: "ranked", GameLabel: "g",
	}, repairSettings())
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 2)
	for _, l := range plan.Lobbies {
		var minR, maxR float64 = 99, 0
		for _, p := range l.Roster {
			minR = min(minR, p.AvgRank)
			maxR = max(maxR, p.AvgRank)
		}
		assert.LessOrEqual(t, maxR-minR, 3.0, "ranked lobbies stay contiguous rank bands")
	}
}
