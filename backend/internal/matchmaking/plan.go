package matchmaking

// PlanEvent runs the full matchmaking pipeline for one game: capacity check, mode strategy,
// sub handling, team split, host pick, and per-lobby fairness evaluation.
func PlanEvent(players []Player, cfg Config, settings Settings) (GamePlan, error) {
	subCount := 0
	for _, p := range players {
		if p.CanSubstitute {
			subCount++
		}
	}

	lobbyCount, err := ValidateCapacity(len(players), subCount, cfg.Slots, cfg.SubMin, cfg.GameLabel)
	if err != nil {
		return GamePlan{}, err
	}
	if lobbyCount == 0 {
		return GamePlan{EventID: cfg.EventID}, nil
	}

	cfg.LobbyCount = lobbyCount

	strategy, ok := strategies[cfg.SortLogic]
	if !ok {
		strategy = AssignBalanced
	}

	lobbies := strategy(players, lobbyCount, cfg.Slots)
	lobbies = ApplySubCapacityRosterConstraint(lobbies, players, lobbyCount, cfg.SubMin)

	for i := range lobbies {
		team1, team2 := SplitIntoTeams(lobbies[i].Roster, cfg.TeamSize)
		lobbies[i].Roster = append(team1, team2...)
	}

	lobbies = AssignMandatorySubs(lobbies, players, lobbyCount, cfg.SubMin)
	lobbies = AssignRemainingAsSubs(lobbies, players)

	for i := range lobbies {
		lobbies[i].HostID = PickLobbyHost(lobbies[i])
		lobbies[i].FairnessWarning = IsLobbyUnfair(lobbies[i], settings, cfg.TierCount)
	}

	return GamePlan{
		EventID: cfg.EventID,
		Lobbies: lobbies,
	}, nil
}

var strategies = map[string]StrategyFunc{
	"balanced": AssignBalanced,
	"ranked":   AssignRanked,
}
