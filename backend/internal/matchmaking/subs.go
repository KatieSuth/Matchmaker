package matchmaking

import "github.com/google/uuid"

// ApplySubCapacityRosterConstraint swaps sub-eligible roster picks for non-subs when mandatory subs require it.
func ApplySubCapacityRosterConstraint(lobbies []LobbyPlan, allPlayers []Player, lobbyCount, subMin int) []LobbyPlan {
	if lobbyCount < 2 {
		return lobbies
	}

	subEligibleTotal := 0
	for _, p := range allPlayers {
		if p.CanSubstitute {
			subEligibleTotal++
		}
	}
	maxOnRoster := MaxSubstituteEligibleOnRoster(subEligibleTotal, lobbyCount, subMin)

	rosteredSubs := collectRosteredSubEligible(lobbies)
	for len(rosteredSubs) > maxOnRoster {
		swapIdx := len(rosteredSubs) - 1
		subPlayer := rosteredSubs[swapIdx]

		replacement, ok := bestNonSubReplacement(allPlayers, lobbies)
		if !ok {
			break
		}

		lobbies = swapRosterPlayer(lobbies, subPlayer.UserID, replacement)
		rosteredSubs = rosteredSubs[:swapIdx]
	}
	return lobbies
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

// bestNonSubReplacement finds the highest-skill unrostered player who cannot substitute.
func bestNonSubReplacement(allPlayers []Player, lobbies []LobbyPlan) (Player, bool) {
	rostered := rosteredUserIDs(lobbies)
	var best *Player
	for _, p := range allPlayers {
		if p.CanSubstitute || rostered[p.UserID] {
			continue
		}
		if best == nil || CompareByRankThenAvailability(p, *best) < 0 {
			cp := p
			best = &cp
		}
	}
	if best == nil {
		return Player{}, false
	}
	return *best, true
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
