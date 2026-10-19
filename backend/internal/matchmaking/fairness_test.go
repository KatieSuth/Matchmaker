package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaledFairnessThresholds(t *testing.T) {
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}
	scaled := matchmaking.ScaledFairnessThresholds(settings, 10)
	assert.InDelta(t, 2.4, scaled.OutlierGap, 0.01)
	assert.InDelta(t, 1.2, scaled.TeamSeparation, 0.01)
}

func TestIsLobbyUnfair_Outlier(t *testing.T) {
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}
	team1 := intPtr(1)
	lobby := matchmaking.LobbyPlan{
		Roster: []matchmaking.Player{
			{UserID: uuid.New(), AvgRank: 25, TeamNumber: team1},
			{UserID: uuid.New(), AvgRank: 12, TeamNumber: team1},
			{UserID: uuid.New(), AvgRank: 11, TeamNumber: team1},
			{UserID: uuid.New(), AvgRank: 10, TeamNumber: team1},
		},
	}
	assert.True(t, matchmaking.IsLobbyUnfair(lobby, settings, 25))
}

func TestIsLobbyUnfair_TeamSeparation(t *testing.T) {
	settings := matchmaking.Settings{
		FairnessOutlierGap:         100,
		FairnessTeamSeparation:     1,
		FairnessReferenceTierCount: 25,
	}
	t1 := intPtr(1)
	t2 := intPtr(2)
	lobby := matchmaking.LobbyPlan{
		Roster: []matchmaking.Player{
			{UserID: uuid.New(), AvgRank: 20, TeamNumber: t1},
			{UserID: uuid.New(), AvgRank: 20, TeamNumber: t1},
			{UserID: uuid.New(), AvgRank: 5, TeamNumber: t2},
			{UserID: uuid.New(), AvgRank: 5, TeamNumber: t2},
		},
	}
	assert.True(t, matchmaking.IsLobbyUnfair(lobby, settings, 25))
}

func TestIsLobbyUnfair_BalancedTeamsNotFlagged(t *testing.T) {
	settings := matchmaking.Settings{
		FairnessOutlierGap:         100,
		FairnessTeamSeparation:     4,
		FairnessReferenceTierCount: 25,
	}
	t1 := intPtr(1)
	t2 := intPtr(2)
	lobby := matchmaking.LobbyPlan{
		Roster: []matchmaking.Player{
			{UserID: uuid.New(), AvgRank: 12, TeamNumber: t1},
			{UserID: uuid.New(), AvgRank: 12, TeamNumber: t1},
			{UserID: uuid.New(), AvgRank: 12, TeamNumber: t2},
			{UserID: uuid.New(), AvgRank: 12, TeamNumber: t2},
		},
	}
	assert.False(t, matchmaking.IsLobbyUnfair(lobby, settings, 25))
}

func TestIsLobbyUnfair_SkipsTeamSeparationWithSingleTeam(t *testing.T) {
	settings := matchmaking.Settings{
		FairnessOutlierGap:         100,
		FairnessTeamSeparation:     0,
		FairnessReferenceTierCount: 25,
	}
	t1 := intPtr(1)
	lobby := matchmaking.LobbyPlan{
		Roster: []matchmaking.Player{
			{UserID: uuid.New(), AvgRank: 20, TeamNumber: t1},
			{UserID: uuid.New(), AvgRank: 5, TeamNumber: t1},
		},
	}
	assert.False(t, matchmaking.IsLobbyUnfair(lobby, settings, 25))
}

func TestTeamAverageSeparationForTest_EmptyTeam(t *testing.T) {
	sep := matchmaking.TeamAverageSeparationForTest(nil, []matchmaking.Player{{AvgRank: 10}})
	assert.Equal(t, 0.0, sep)
}

func TestOutlierGapForTest(t *testing.T) {
	assert.Equal(t, 0.0, matchmaking.OutlierGapForTest(nil))
	assert.Equal(t, 0.0, matchmaking.OutlierGapForTest([]matchmaking.Player{{AvgRank: 13}}))
	assert.Equal(t, 7.0, matchmaking.OutlierGapForTest([]matchmaking.Player{
		{AvgRank: 13}, {AvgRank: 6}, {AvgRank: 6}, {AvgRank: 3},
	}))
}

func TestSplitRosterByTeamNumberForTest(t *testing.T) {
	t1 := intPtr(1)
	t2 := intPtr(2)
	team1, team2 := matchmaking.SplitRosterByTeamNumberForTest([]matchmaking.Player{
		{UserID: uuid.New(), TeamNumber: t1, AvgRank: 10},
		{UserID: uuid.New(), TeamNumber: t2, AvgRank: 8},
		{UserID: uuid.New(), TeamNumber: nil, AvgRank: 5},
	})
	require.Len(t, team1, 1)
	require.Len(t, team2, 1)
}

func TestScaledFairnessThresholds_ZeroTierCountUsesBaseline(t *testing.T) {
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}
	scaled := matchmaking.ScaledFairnessThresholds(settings, 0)
	assert.Equal(t, 6.0, scaled.OutlierGap)
	assert.Equal(t, 3.0, scaled.TeamSeparation)
}

func intPtr(v int) *int {
	return &v
}

func TestEvaluateLobbyFairness_ReportsEachCheck(t *testing.T) {
	settings := matchmaking.Settings{FairnessOutlierGap: 6, FairnessTeamSeparation: 3, FairnessReferenceTierCount: 25}
	lobby := func(team1, team2 []float64) matchmaking.LobbyPlan {
		var roster []matchmaking.Player
		for _, r := range team1 {
			roster = append(roster, matchmaking.Player{UserID: uuid.New(), AvgRank: r, TeamNumber: intPtr(1)})
		}
		for _, r := range team2 {
			roster = append(roster, matchmaking.Player{UserID: uuid.New(), AvgRank: r, TeamNumber: intPtr(2)})
		}
		return matchmaking.LobbyPlan{Roster: roster}
	}

	cases := []struct {
		name         string
		team1, team2 []float64
		want         matchmaking.FairnessChecks
	}{
		{"fair", []float64{10, 9}, []float64{10, 9}, matchmaking.FairnessChecks{}},
		{"outlier only", []float64{20, 4}, []float64{10, 10}, matchmaking.FairnessChecks{Outlier: true}},
		{"team separation only", []float64{14, 13}, []float64{9, 9}, matchmaking.FairnessChecks{TeamSeparation: true}},
		{"both", []float64{25, 10}, []float64{5, 5}, matchmaking.FairnessChecks{Outlier: true, TeamSeparation: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := lobby(tc.team1, tc.team2)
			got := matchmaking.EvaluateLobbyFairness(l, settings, 25)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, got.Unfair(), matchmaking.IsLobbyUnfair(l, settings, 25))
		})
	}
}

func TestPlanEvent_RecordsFairnessChecksWithWarning(t *testing.T) {
	now := time.Now()
	var players []matchmaking.Player
	for i, r := range []float64{25, 3, 3, 3} {
		players = append(players, matchmaking.Player{UserID: uuid.New(), AvgRank: r, CreatedAt: now.Add(time.Duration(i) * time.Second)})
	}
	plan, err := matchmaking.PlanEvent(players, matchmaking.Config{
		EventID: uuid.New(), TeamSize: 2, Slots: 4, TierCount: 25, SortLogic: "balanced", GameLabel: "g",
	}, matchmaking.Settings{FairnessOutlierGap: 6, FairnessTeamSeparation: 3, FairnessReferenceTierCount: 25})
	require.NoError(t, err)
	lobby := plan.Lobbies[0]
	assert.True(t, lobby.FairnessWarning)
	assert.True(t, lobby.FairnessChecks.Outlier)
	assert.True(t, lobby.FairnessChecks.TeamSeparation)
}
