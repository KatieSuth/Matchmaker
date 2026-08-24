package matchmaking

import (
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

// outlierGap is highest roster rank minus the next highest. Empty or single-player
// rosters have no gap.
func outlierGap(roster []Player) float64 {
	if len(roster) < 2 {
		return 0
	}
	ranks := make([]float64, len(roster))
	for i, p := range roster {
		ranks[i] = p.AvgRank
	}
	sort.Float64s(ranks)
	return ranks[len(ranks)-1] - ranks[len(ranks)-2]
}

// outlierExceeds is true when the top player is much higher ranked than the rest of the lobby.
func outlierExceeds(roster []Player, gap float64) bool {
	return outlierGap(roster) > gap
}

// teamSeparationExceeds is true when the two teams' average skill differs beyond the threshold.
func teamSeparationExceeds(lobby LobbyPlan, separation float64) bool {
	team1, team2 := splitRosterByTeamNumber(lobby.Roster)
	if len(team1) == 0 || len(team2) == 0 {
		return false
	}
	return teamAverageSeparation(team1, team2) > separation
}

// splitRosterByTeamNumber partitions a lobby roster into team 1 and team 2 players.
func splitRosterByTeamNumber(roster []Player) ([]Player, []Player) {
	var team1, team2 []Player
	for _, p := range roster {
		if p.TeamNumber == nil {
			continue
		}
		switch *p.TeamNumber {
		case 1:
			team1 = append(team1, p)
		case 2:
			team2 = append(team2, p)
		}
	}
	return team1, team2
}
