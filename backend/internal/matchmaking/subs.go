package matchmaking

import "github.com/google/uuid"

// ApplySubCapacityRosterConstraint swaps sub-eligible roster picks for non-subs when mandatory subs require it.
// Replacements stay in the vacated rank window so a Gold hole does not get a leftover Radiant.
func ApplySubCapacityRosterConstraint(lobbies []LobbyPlan, allPlayers []Player, cfg Config) ([]LobbyPlan, bool) {
	if cfg.LobbyCount < 2 {
		return lobbies, false
	}

	subEligibleTotal := 0
	for _, p := range allPlayers {
		if p.CanSubstitute {
			subEligibleTotal++
		}
	}
	maxOnRoster := MaxSubstituteEligibleOnRoster(subEligibleTotal, cfg.LobbyCount, cfg.SubMin)
	windows := buildRankWindows(cfg.TierCount, cfg.TeamSize)

	adjusted := false
	rosteredSubs := collectRosteredSubEligible(lobbies)
	for len(rosteredSubs) > maxOnRoster {
		swapIdx := len(rosteredSubs) - 1
		subPlayer := rosteredSubs[swapIdx]

		replacement, ok := bestNonSubReplacement(allPlayers, lobbies, subPlayer, windows, cfg.TierCount)
		if !ok {
			break
		}

		lobbies = swapRosterPlayer(lobbies, subPlayer.UserID, replacement)
		rosteredSubs = rosteredSubs[:swapIdx]
		adjusted = true
	}
	return lobbies, adjusted
}

// collectRosteredSubEligible returns sub-eligible players currently on a team roster, highest skill first.
func collectRosteredSubEligible(lobbies []LobbyPlan) []Player {
	var out []Player
	for _, lobby := range lobbies {
		for _, p := range lobby.Roster {
			if p.CanSubstitute {
				out = append(out, p)
			}
		}
	}
	sortPlayersByRankDesc(out)
	return out
}

// bestNonSubReplacement finds an unrostered non-sub to fill the vacated slot.
// Same rank window first, then the nearest window; within that band, closest to the
// vacated rank. Highest leftover skill is not used — that is how Radiant lands in a Gold hole.
func bestNonSubReplacement(allPlayers []Player, lobbies []LobbyPlan, vacated Player, windows []rankWindow, tierCount int) (Player, bool) {
	rostered := rosteredUserIDs(lobbies)
	var candidates []Player
	for _, p := range allPlayers {
		if p.CanSubstitute || rostered[p.UserID] {
			continue
		}
		candidates = append(candidates, p)
	}
	if len(candidates) == 0 {
		return Player{}, false
	}
	if len(windows) == 0 {
		return pickCloserToTarget(candidates, vacated.AvgRank), true
	}

	vacatedIdx := windowIndexForOrder(clampedRankOrder(vacated.AvgRank, tierCount), windows)
	for dist := 0; dist < len(windows); dist++ {
		var band []Player
		for _, p := range candidates {
			idx := windowIndexForOrder(clampedRankOrder(p.AvgRank, tierCount), windows)
			if absInt(idx-vacatedIdx) == dist {
				band = append(band, p)
			}
		}
		if len(band) > 0 {
			return pickCloserToTarget(band, vacated.AvgRank), true
		}
	}
	return pickCloserToTarget(candidates, vacated.AvgRank), true
}

// pickCloserToTarget returns the player nearest target skill; ties use CompareByRankThenAvailability.
func pickCloserToTarget(players []Player, target float64) Player {
	if len(players) == 0 {
		return Player{}
	}
	best := players[0]
	for _, p := range players[1:] {
		if closerToMidpoint(p, best, target) {
			best = p
		}
	}
	return best
}

// rosteredUserIDs lists every player already assigned to a lobby roster or sub pool.
func rosteredUserIDs(lobbies []LobbyPlan) map[uuid.UUID]bool {
	ids := make(map[uuid.UUID]bool)
	for _, lobby := range lobbies {
		for _, p := range lobby.Roster {
			ids[p.UserID] = true
		}
		for _, p := range lobby.Subs {
			ids[p.UserID] = true
		}
	}
	return ids
}

// swapRosterPlayer replaces one rostered player with another across all lobbies.
func swapRosterPlayer(lobbies []LobbyPlan, removeID uuid.UUID, add Player) []LobbyPlan {
	for i := range lobbies {
		for j := range lobbies[i].Roster {
			if lobbies[i].Roster[j].UserID == removeID {
				lobbies[i].Roster[j] = add
				return lobbies
			}
		}
	}
	return lobbies
}

// AssignMandatorySubs places sub_min subs per lobby from unassigned sub-eligible players.
func AssignMandatorySubs(lobbies []LobbyPlan, allPlayers []Player, lobbyCount, subMin int) []LobbyPlan {
	if lobbyCount < 2 || subMin == 0 {
		return lobbies
	}

	assigned := rosteredUserIDs(lobbies)
	var available []Player
	for _, p := range allPlayers {
		if p.CanSubstitute && !assigned[p.UserID] {
			available = append(available, p)
		}
	}
	sortPlayersByRankDesc(available)

	idx := 0
	for i := range lobbies {
		for s := 0; s < subMin && idx < len(available); s++ {
			sub := available[idx]
			sub.TeamNumber = nil
			lobbies[i].Subs = append(lobbies[i].Subs, sub)
			assigned[sub.UserID] = true
			idx++
		}
	}
	return lobbies
}

// AssignRemainingAsSubs distributes overflow sub-eligible players round-robin across lobbies.
func AssignRemainingAsSubs(lobbies []LobbyPlan, allPlayers []Player) []LobbyPlan {
	if len(lobbies) == 0 {
		return lobbies
	}

	assigned := rosteredUserIDs(lobbies)
	var remaining []Player
	for _, p := range allPlayers {
		if p.CanSubstitute && !assigned[p.UserID] {
			remaining = append(remaining, p)
		}
	}
	sortPlayersByRankDesc(remaining)

	for i, p := range remaining {
		p.TeamNumber = nil
		lobbyIdx := i % len(lobbies)
		lobbies[lobbyIdx].Subs = append(lobbies[lobbyIdx].Subs, p)
	}
	return lobbies
}
