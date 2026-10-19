package matchmaking

// PlanEvent runs the full matchmaking pipeline for one game: capacity check, mode strategy,
// sub handling, team split, balanced fairness repair, host pick, and per-lobby fairness evaluation.
// Lobby count is always sized from every registrant, so no-show deprioritization never reduces it.
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
	plan := buildGamePlan(players, cfg, settings)

	// MaxLobbies is a count ceiling, including n that benches every can-sub.
	// Keep that n when the non-sub rosters are even; drop only while they stay unfair.
	for shouldDropReservedSubLobby(subCount, cfg) && anyLobbyUnfair(plan.Lobbies) {
		cfg.LobbyCount--
		plan = buildGamePlan(players, cfg, settings)
	}

	return plan, nil
}

// buildGamePlan runs the pipeline for one lobby count. PlanEvent may call it
// again at n−1 when a reserved-can-sub extra lobby is still unfair. Deprioritized
// no-shows are left out of the lobby fill and sub-capacity replacements but stay in
// the pool for fairness repair (as a last resort) and for sub assignment.
func buildGamePlan(players []Player, cfg Config, settings Settings) GamePlan {
	lobbyCount := cfg.LobbyCount
	// Repeat no-shows sit out the fill; add-back depends on n, so this reruns when n drops.
	players = markDeprioritized(players, cfg)
	fillPool := eligiblePlayers(players)

	var lobbies []LobbyPlan
	if cfg.SortLogic == "ranked" {
		lobbies = AssignRanked(fillPool, lobbyCount, cfg.Slots)
	} else {
		lobbies = AssignBalanced(fillPool, cfg)
	}
	lobbies, subCapacityAdjusted := ApplySubCapacityRosterConstraint(lobbies, players, cfg)
	lobbies = ApplyDuoLobbyGrouping(lobbies)

	for i := range lobbies {
		var team1, team2 []Player
		if cfg.SortLogic == "ranked" {
			team1, team2 = SplitIntoTeams(lobbies[i].Roster, cfg.TeamSize)
		} else {
			team1, team2 = splitIntoTeamsWindowed(lobbies[i].Roster, cfg.TeamSize, cfg.TierCount)
		}
		lobbies[i].Roster = append(team1, team2...)
	}

	// Unplaced leftovers can repair an unfair split (lopsided teams or a rank outlier) that
	// fill + sub-capacity never reconsidered. Ranked keeps its packed band and only takes a
	// single swap that makes the lobby fully fair.
	if cfg.SortLogic == "ranked" {
		lobbies = repairUnfairRankedLobbies(lobbies, players, cfg, settings)
	} else {
		// Closest-average teams face each other; re-checked after repair, which can move averages.
		lobbies = pairTeamsAcrossLobbies(lobbies, settings, cfg.TierCount)
		lobbies = repairUnfairWindowPairs(lobbies, players, cfg, settings)
		lobbies = pairTeamsAcrossLobbies(lobbies, settings, cfg.TierCount)
	}

	lobbies = AssignMandatorySubs(lobbies, players, lobbyCount, cfg.SubMin)
	lobbies = AssignRemainingAsSubs(lobbies, players)

	for i := range lobbies {
		lobbies[i].HostID = PickLobbyHost(lobbies[i])
		lobbies[i].FairnessChecks = EvaluateLobbyFairness(lobbies[i], settings, cfg.TierCount)
		lobbies[i].FairnessWarning = lobbies[i].FairnessChecks.Unfair()
	}

	// Host warning is for leftover imbalance from the sub_min swap, not the swap itself.
	if subCapacityAdjusted && !anyLobbyUnfair(lobbies) {
		subCapacityAdjusted = false
	}

	return GamePlan{
		EventID:             cfg.EventID,
		Lobbies:             lobbies,
		SubCapacityAdjusted: subCapacityAdjusted,
	}
}

// shouldDropReservedSubLobby is true when this n benches every can-sub. PlanEvent
// then retries n−1 only if the current plan is still unfair.
func shouldDropReservedSubLobby(subCount int, cfg Config) bool {
	if cfg.LobbyCount < 2 || cfg.SubMin <= 0 {
		return false
	}
	return MaxSubstituteEligibleOnRoster(subCount, cfg.LobbyCount, cfg.SubMin) <= 0
}

// anyLobbyUnfair is true when at least one lobby already has FairnessWarning set.
// Callers must run after host/fairness assignment; it does not recompute checks.
func anyLobbyUnfair(lobbies []LobbyPlan) bool {
	for _, lobby := range lobbies {
		if lobby.FairnessWarning {
			return true
		}
	}
	return false
}
