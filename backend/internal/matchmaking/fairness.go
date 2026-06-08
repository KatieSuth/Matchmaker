package matchmaking

import (
	"math"
	"sort"
)

// ScaledThresholds holds tier-adjusted fairness cutoffs.
type ScaledThresholds struct {
	OutlierGap     float64
	TeamSeparation float64
}

// ScaledFairnessThresholds scales env baselines to the game's tier count.
func ScaledFairnessThresholds(settings Settings, tierCount int) ScaledThresholds {
	if tierCount <= 0 || settings.FairnessReferenceTierCount <= 0 {
		return ScaledThresholds{
			OutlierGap:     float64(settings.FairnessOutlierGap),
			TeamSeparation: settings.FairnessTeamSeparation,
		}
	}
	scale := float64(tierCount) / float64(settings.FairnessReferenceTierCount)
	return ScaledThresholds{
		OutlierGap:     float64(settings.FairnessOutlierGap) * scale,
		TeamSeparation: settings.FairnessTeamSeparation * scale,
	}
}

// IsLobbyUnfair flags a lobby when outlier or team-average separation exceeds scaled thresholds.
func IsLobbyUnfair(lobby LobbyPlan, settings Settings, tierCount int) bool {
	thresholds := ScaledFairnessThresholds(settings, tierCount)
	if outlierExceeds(lobby.Roster, thresholds.OutlierGap) {
		return true
	}
	return teamSeparationExceeds(lobby, thresholds.TeamSeparation)
}

// outlierExceeds is true when the top player is much higher ranked than the rest of the lobby.
func outlierExceeds(roster []Player, gap float64) bool {
	if len(roster) < 2 {
		return false
	}
	ranks := make([]float64, len(roster))
	for i, p := range roster {
		ranks[i] = p.AvgRank
	}
	sort.Float64s(ranks)
	highest := ranks[len(ranks)-1]
	second := ranks[len(ranks)-2]
	return highest-second > gap
}

// teamSeparationExceeds is true when the two teams' average skill differs beyond the threshold.
func teamSeparationExceeds(lobby LobbyPlan, separation float64) bool {
	var team1Sum, team2Sum float64
	var team1Count, team2Count int

	for _, p := range lobby.Roster {
		if p.TeamNumber == nil {
			continue
		}
		switch *p.TeamNumber {
		case 1:
			team1Sum += p.AvgRank
			team1Count++
		case 2:
			team2Sum += p.AvgRank
			team2Count++
		}
	}

	if team1Count == 0 || team2Count == 0 {
		return false
	}

	team1Avg := team1Sum / float64(team1Count)
	team2Avg := team2Sum / float64(team2Count)
	return math.Abs(team1Avg-team2Avg) > separation
}
