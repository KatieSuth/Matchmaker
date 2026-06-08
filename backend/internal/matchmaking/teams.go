package matchmaking

// SplitIntoTeams assigns team_number 1 or 2 via snake draft within a lobby roster,
// then attempts to honor mutual duo requests without worsening team balance.
func SplitIntoTeams(roster []Player, teamSize int) ([]Player, []Player) {
	team1, team2 := snakeDraftIntoTeams(roster, teamSize)
	if len(team1) == 0 && len(team2) == 0 {
		return team1, team2
	}

	baselineSep := teamAverageSeparation(team1, team2)
	team1, team2 = applyDuoTeamGrouping(team1, team2, baselineSep)
	return team1, team2
}

// snakeDraftIntoTeams assigns team_number 1 or 2 via snake draft without duo adjustments.
func snakeDraftIntoTeams(roster []Player, teamSize int) ([]Player, []Player) {
	if len(roster) == 0 {
		return nil, nil
	}
	sorted := append([]Player(nil), roster...)
	sortPlayersByRankDesc(sorted)

	team1 := make([]Player, 0, teamSize)
	team2 := make([]Player, 0, teamSize)

	for i, p := range sorted {
		teamNum := 1
		round := i / 2
		posInRound := i % 2
		if round%2 == 0 {
			if posInRound == 1 {
				teamNum = 2
			}
		} else {
			if posInRound == 0 {
				teamNum = 2
			}
		}
		p.TeamNumber = intPtr(teamNum)
		if teamNum == 1 {
			team1 = append(team1, p)
		} else {
			team2 = append(team2, p)
		}
	}
	return team1, team2
}

// sortPlayersByRankDesc orders players highest skill first, using availability as a tie-break.
func sortPlayersByRankDesc(players []Player) {
	for i := 1; i < len(players); i++ {
		for j := i; j > 0 && CompareByRankThenAvailability(players[j], players[j-1]) < 0; j-- {
			players[j], players[j-1] = players[j-1], players[j]
		}
	}
}

func intPtr(v int) *int {
	return &v
}
