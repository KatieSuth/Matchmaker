package matchmaking_test

import (
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaledFairnessThresholds(t *testing.T) {
	settings := matchmaking.Settings{
		FairnessOutlierGap:         8,
		FairnessTeamSeparation:     4,
		FairnessReferenceTierCount: 25,
	}
	scaled := matchmaking.ScaledFairnessThresholds(settings, 10)
	assert.InDelta(t, 3.2, scaled.OutlierGap, 0.01)
	assert.InDelta(t, 1.6, scaled.TeamSeparation, 0.01)
}

func TestIsLobbyUnfair_Outlier(t *testing.T) {
	settings := matchmaking.Settings{
		FairnessOutlierGap:         8,
		FairnessTeamSeparation:     4,
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
		FairnessOutlierGap:         8,
		FairnessTeamSeparation:     4,
		FairnessReferenceTierCount: 25,
	}
	scaled := matchmaking.ScaledFairnessThresholds(settings, 0)
	assert.Equal(t, 8.0, scaled.OutlierGap)
	assert.Equal(t, 4.0, scaled.TeamSeparation)
}

func intPtr(v int) *int {
	return &v
}
