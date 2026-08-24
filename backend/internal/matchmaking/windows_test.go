package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRankWindows_ValorantFiveVFive(t *testing.T) {
	windows := matchmaking.BuildRankWindowsForTest(25, 5)
	require.Len(t, windows, 5)
	assert.Equal(t, 1, windows[0].MinOrder)
	assert.Equal(t, 5, windows[0].MaxOrder)
	assert.Equal(t, 6, windows[1].MinOrder)
	assert.Equal(t, 10, windows[1].MaxOrder)
	assert.Equal(t, 11, windows[2].MinOrder)
	assert.Equal(t, 15, windows[2].MaxOrder)
	assert.Equal(t, 16, windows[3].MinOrder)
	assert.Equal(t, 20, windows[3].MaxOrder)
	assert.Equal(t, 21, windows[4].MinOrder)
	assert.Equal(t, 25, windows[4].MaxOrder)
}

func TestBuildRankWindows_ValorantThreeVThreeWeightsLow(t *testing.T) {
	windows := matchmaking.BuildRankWindowsForTest(25, 3)
	require.Len(t, windows, 3)
	assert.Equal(t, 1, windows[0].MinOrder)
	assert.Equal(t, 9, windows[0].MaxOrder)
	assert.Equal(t, 10, windows[1].MinOrder)
	assert.Equal(t, 17, windows[1].MaxOrder)
	assert.Equal(t, 18, windows[2].MinOrder)
	assert.Equal(t, 25, windows[2].MaxOrder)
}

func TestBuildRankWindows_ValorantFourVFourWeightsLow(t *testing.T) {
	windows := matchmaking.BuildRankWindowsForTest(25, 4)
	require.Len(t, windows, 4)
	assert.Equal(t, 7, windows[0].MaxOrder-windows[0].MinOrder+1)
	assert.Equal(t, 6, windows[1].MaxOrder-windows[1].MinOrder+1)
	assert.Equal(t, 6, windows[2].MaxOrder-windows[2].MinOrder+1)
	assert.Equal(t, 6, windows[3].MaxOrder-windows[3].MinOrder+1)
}

func TestBuildRankWindows_ShrinksWhenFewerTiersThanTeamSize(t *testing.T) {
	windows := matchmaking.BuildRankWindowsForTest(4, 5)
	require.Len(t, windows, 4)
	assert.Equal(t, 1, windows[0].MinOrder)
	assert.Equal(t, 1, windows[0].MaxOrder)
	assert.Equal(t, 4, windows[3].MaxOrder)
}

func TestBuildRankWindows_InvalidInputEmpty(t *testing.T) {
	assert.Empty(t, matchmaking.BuildRankWindowsForTest(0, 5))
	assert.Empty(t, matchmaking.BuildRankWindowsForTest(25, 0))
}

func TestClampedRankOrder_Bounds(t *testing.T) {
	assert.Equal(t, 1, matchmaking.ClampedRankOrderForTest(0, 25))
	assert.Equal(t, 1, matchmaking.ClampedRankOrderForTest(-3, 25))
	assert.Equal(t, 25, matchmaking.ClampedRankOrderForTest(40, 25))
	assert.Equal(t, 13, matchmaking.ClampedRankOrderForTest(13, 25))
}

func TestPickClosestToMidpoint_PrefersNearerRanks(t *testing.T) {
	now := time.Now()
	low := playerAtRank(1, false, 1, now)
	mid := playerAtRank(8, false, 1, now.Add(time.Minute))
	high := playerAtRank(13, false, 1, now.Add(2*time.Minute))
	picked, rest := matchmaking.PickClosestToMidpointForTest([]matchmaking.Player{high, low, mid}, 2, 7)
	require.Len(t, picked, 2)
	require.Len(t, rest, 1)
	ids := map[uuid.UUID]bool{picked[0].UserID: true, picked[1].UserID: true}
	assert.True(t, ids[mid.UserID])
	assert.True(t, ids[high.UserID])
	assert.Equal(t, low.UserID, rest[0].UserID)
}

func TestFullestWindowIndex_PrefersLowerOnTie(t *testing.T) {
	a := playerAtRank(2, false, 1, time.Now())
	b := playerAtRank(12, false, 1, time.Now())
	remaining := [][]matchmaking.Player{
		{a, a},
		{b, b},
		{},
	}
	assert.Equal(t, 0, matchmaking.FullestWindowIndexForTest(remaining))
}

func TestWindowDraftIntoTeams_EmptyRoster(t *testing.T) {
	team1, team2 := matchmaking.WindowDraftIntoTeamsForTest(nil, 5, 25)
	assert.Nil(t, team1)
	assert.Nil(t, team2)
}

func TestWindowDraftIntoTeams_FallsBackToSnakeWhenLadderInvalid(t *testing.T) {
	now := time.Now()
	roster := []matchmaking.Player{
		{UserID: uuid.New(), AvgRank: 20, CreatedAt: now},
		{UserID: uuid.New(), AvgRank: 18, CreatedAt: now.Add(time.Minute)},
		{UserID: uuid.New(), AvgRank: 16, CreatedAt: now.Add(2 * time.Minute)},
		{UserID: uuid.New(), AvgRank: 14, CreatedAt: now.Add(3 * time.Minute)},
	}
	windowed1, windowed2 := matchmaking.WindowDraftIntoTeamsForTest(roster, 2, 0)
	snake1, snake2 := matchmaking.SnakeDraftIntoTeamsForTest(roster, 2)
	require.Len(t, windowed1, 2)
	require.Len(t, snake1, 2)
	assert.Equal(t, *snake1[0].TeamNumber, *windowed1[0].TeamNumber)
	assert.Equal(t, snake2[0].AvgRank, windowed2[0].AvgRank)
}

func TestWindowDraftIntoTeams_OneWindowUsesAdjacentPairsNotSnake(t *testing.T) {
	now := time.Now()
	ids := make([]uuid.UUID, 6)
	roster := make([]matchmaking.Player, 6)
	ranks := []float64{20, 12, 11, 10, 9, 4}
	for i, rank := range ranks {
		ids[i] = uuid.New()
		roster[i] = matchmaking.Player{UserID: ids[i], AvgRank: rank, CreatedAt: now.Add(time.Duration(i) * time.Minute)}
	}

	team1, team2 := matchmaking.WindowDraftIntoTeamsForTest(roster, 3, 25)
	require.Len(t, team1, 3)
	require.Len(t, team2, 3)

	sum1, sum2 := 0.0, 0.0
	for _, p := range team1 {
		sum1 += p.AvgRank
	}
	for _, p := range team2 {
		sum2 += p.AvgRank
	}
	// Snake would be 20+10+9 vs 12+11+4 (39 vs 27). Adjacent-to-weaker is much closer.
	assert.InDelta(t, sum1, sum2, 6)
	assert.Less(t, matchmaking.TeamAverageSeparationForTest(team1, team2), 13.0)
}

func TestWindowDraftIntoTeams_ThreeVThreeClusteredGoldsAndLowPair(t *testing.T) {
	now := time.Now()
	golds := make([]uuid.UUID, 4)
	roster := make([]matchmaking.Player, 0, 6)
	for i, rank := range []float64{12, 11, 10, 10} {
		golds[i] = uuid.New()
		roster = append(roster, matchmaking.Player{UserID: golds[i], AvgRank: rank, CreatedAt: now.Add(time.Duration(i) * time.Minute)})
	}
	silver := uuid.New()
	iron := uuid.New()
	roster = append(roster,
		matchmaking.Player{UserID: silver, AvgRank: 8, CreatedAt: now.Add(4 * time.Minute)},
		matchmaking.Player{UserID: iron, AvgRank: 1, CreatedAt: now.Add(5 * time.Minute)},
	)

	team1, team2 := matchmaking.WindowDraftIntoTeamsForTest(roster, 3, 25)
	teamOf := func(id uuid.UUID) int {
		for _, p := range team1 {
			if p.UserID == id {
				return 1
			}
		}
		for _, p := range team2 {
			if p.UserID == id {
				return 2
			}
		}
		return 0
	}
	assert.NotEqual(t, teamOf(silver), teamOf(iron), "low-window pair should be on opposite teams")

	goldsOn := func(team []matchmaking.Player) int {
		n := 0
		for _, p := range team {
			for _, id := range golds {
				if p.UserID == id {
					n++
				}
			}
		}
		return n
	}
	assert.Equal(t, 2, goldsOn(team1))
	assert.Equal(t, 2, goldsOn(team2))
}

func TestWindowIndexForOrder_BoundsAndGaps(t *testing.T) {
	assert.Equal(t, 0, matchmaking.WindowIndexForOrderForTest(5, nil))

	windows := []matchmaking.RankWindowForTest{
		{MinOrder: 1, MaxOrder: 5, Midpoint: 3},
		{MinOrder: 11, MaxOrder: 15, Midpoint: 13},
	}
	assert.Equal(t, 0, matchmaking.WindowIndexForOrderForTest(0, windows))
	assert.Equal(t, 0, matchmaking.WindowIndexForOrderForTest(3, windows))
	assert.Equal(t, 1, matchmaking.WindowIndexForOrderForTest(40, windows))
	assert.Equal(t, 1, matchmaking.WindowIndexForOrderForTest(8, windows), "gap falls through to last window")
}

func TestBestLeftoverPair_CloserDistanceAndAdjacentTie(t *testing.T) {
	now := time.Now()
	closer := []matchmaking.WindowLeftoverForTest{
		{Player: playerAtRank(25, false, 1, now), Window: 2, Order: 25},
		{Player: playerAtRank(12, false, 1, now), Window: 1, Order: 12},
		{Player: playerAtRank(1, false, 1, now), Window: 0, Order: 1},
	}
	i, j := matchmaking.BestLeftoverPairFromTest(closer)
	assert.Equal(t, 1, i)
	assert.Equal(t, 2, j)

	tied := []matchmaking.WindowLeftoverForTest{
		{Player: playerAtRank(5, false, 1, now), Window: 0, Order: 5},
		{Player: playerAtRank(15, false, 1, now), Window: 2, Order: 15},
		{Player: playerAtRank(25, false, 1, now), Window: 3, Order: 25},
	}
	i, j = matchmaking.BestLeftoverPairFromTest(tied)
	assert.Equal(t, 1, i)
	assert.Equal(t, 2, j, "equal distance prefers adjacent windows")
}

func TestRemoveLeftoverPair_SwapsWhenFirstIndexIsLarger(t *testing.T) {
	now := time.Now()
	items := []matchmaking.WindowLeftoverForTest{
		{Player: playerAtRank(1, false, 1, now), Window: 0, Order: 1},
		{Player: playerAtRank(10, false, 1, now), Window: 1, Order: 10},
		{Player: playerAtRank(20, false, 1, now), Window: 2, Order: 20},
	}
	out := matchmaking.RemoveLeftoverPairFromTest(items, 2, 0)
	require.Len(t, out, 1)
	assert.Equal(t, 10, out[0].Order)
}

func TestAppendToWeakerTeam_BothSides(t *testing.T) {
	now := time.Now()
	weak := playerAtRank(8, false, 1, now)
	strong := playerAtRank(20, false, 1, now.Add(time.Minute))

	var team1, team2 []matchmaking.Player
	sum1, sum2 := 0.0, 5.0
	matchmaking.AppendToWeakerTeamForTest(weak, &team1, &team2, &sum1, &sum2)
	require.Len(t, team1, 1)
	assert.Equal(t, 8.0, sum1)
	require.NotNil(t, team1[0].TeamNumber)
	assert.Equal(t, 1, *team1[0].TeamNumber)

	sum1, sum2 = 10.0, 0.0
	team1, team2 = nil, nil
	matchmaking.AppendToWeakerTeamForTest(strong, &team1, &team2, &sum1, &sum2)
	require.Len(t, team2, 1)
	assert.Equal(t, 20.0, sum2)
	require.NotNil(t, team2[0].TeamNumber)
	assert.Equal(t, 2, *team2[0].TeamNumber)
}

func TestWindowDraftIntoTeams_OddLeftoverGoesToWeakerTeam(t *testing.T) {
	now := time.Now()
	roster := []matchmaking.Player{
		{UserID: uuid.New(), AvgRank: 21, CreatedAt: now},
		{UserID: uuid.New(), AvgRank: 21, CreatedAt: now.Add(time.Minute)},
		{UserID: uuid.New(), AvgRank: 8, CreatedAt: now.Add(2 * time.Minute)},
	}
	team1, team2 := matchmaking.WindowDraftIntoTeamsForTest(roster, 2, 25)
	assert.Equal(t, 3, len(team1)+len(team2))
	assert.True(t, len(team1) == 2 || len(team2) == 2)
}
