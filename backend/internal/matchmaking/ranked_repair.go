package matchmaking

import "math"

// repairUnfairRankedLobbies is ranked mode's final fairness check for no-show deprioritization.
// Ranked mode otherwise keeps its packed rank band and never swaps leftovers in for balance.
// But a held-out no-show sat out the packing only because of their absences, so when one of
// them could make an unfair lobby fully fair with a single swap, the most balanced such swap is
// made; a clean leftover wins only when it balances just as well. Without deprioritized leftovers
// this is a no-op.
// Other lobbies' players and reserved can-subs are never taken.
func repairUnfairRankedLobbies(lobbies []LobbyPlan, allPlayers []Player, cfg Config, settings Settings) []LobbyPlan {
	reserved := reservedCanSubs(cfg)
	for i := range lobbies {
		if !IsLobbyUnfair(lobbies[i], settings, cfg.TierCount) {
			continue
		}
		leftovers := unplacedPlayers(allPlayers, lobbies)
		unrosteredCanSubs := countUnrosteredCanSubs(allPlayers, lobbies)
		roster, usesDeprioritized, ok := bestFairRankedSwap(lobbies[i], leftovers, unrosteredCanSubs, reserved, cfg, settings)
		if !ok {
			continue
		}
		// Only swap when a held-out player could fix the lobby; a clean-only fix would break
		// ranked's band without any deprioritized player having been passed over.
		if !usesDeprioritized && !anyFairDeprioritizedSwap(lobbies[i], leftovers, unrosteredCanSubs, reserved, cfg, settings) {
			continue
		}
		lobbies[i].Roster = roster
	}
	return lobbies
}

// bestFairRankedSwap returns the most balanced roster reachable by one leftover-for-seat swap
// (seat kept, or the lobby re-drafted), whether it brings in a deprioritized player, and false
// when no swap leaves the lobby fair. Most even teams win, then the closer rank swap (least change
// to the lobby's band); a clean leftover beats a deprioritized one only when it balances as well.
func bestFairRankedSwap(lobby LobbyPlan, leftovers []Player, unrosteredCanSubs, reserved int, cfg Config, settings Settings) ([]Player, bool, bool) {
	before := captureRepairState(lobby, settings, cfg.TierCount)
	overall := inPlaceBest{state: before, delta: math.Inf(1)}
	clean := inPlaceBest{state: before, delta: math.Inf(1)}

	for _, leftover := range leftovers {
		if !leftoverCanJoinRoster(leftover, unrosteredCanSubs, reserved) {
			continue
		}
		for _, seat := range lobby.Roster {
			if seat.TeamNumber == nil {
				continue
			}
			inPlace := replaceRosterPlayerKeepingTeam(lobby.Roster, seat.UserID, leftover)
			team1, team2 := SplitIntoTeams(clearTeamNumbers(inPlace), cfg.TeamSize)
			redrafted := append(team1, team2...)
			delta := math.Abs(seat.AvgRank - leftover.AvgRank)

			for _, trial := range [][]Player{inPlace, redrafted} {
				state := captureRepairState(LobbyPlan{Roster: trial}, settings, cfg.TierCount)
				if state.sepOver || state.outOver {
					continue
				}
				overall.offer(true, state, delta, trial, seat, leftover)
				if !leftover.Deprioritized {
					clean.offer(true, state, delta, trial, seat, leftover)
				}
			}
		}
	}

	if !overall.found {
		return nil, false, false
	}
	if clean.found && !moreBalanced(repairCandidate{state: overall.state}, repairCandidate{state: clean.state}) {
		return clean.roster, false, true
	}
	return overall.roster, overall.in.Deprioritized, true
}

// anyFairDeprioritizedSwap reports whether some deprioritized leftover alone could make the lobby fair.
func anyFairDeprioritizedSwap(lobby LobbyPlan, leftovers []Player, unrosteredCanSubs, reserved int, cfg Config, settings Settings) bool {
	var held []Player
	for _, p := range leftovers {
		if p.Deprioritized {
			held = append(held, p)
		}
	}
	if len(held) == 0 {
		return false
	}
	_, _, ok := bestFairRankedSwap(lobby, held, unrosteredCanSubs, reserved, cfg, settings)
	return ok
}

// clearTeamNumbers copies roster with team assignments removed so it can be re-drafted.
func clearTeamNumbers(roster []Player) []Player {
	out := make([]Player, len(roster))
	for i, p := range roster {
		p.TeamNumber = nil
		out[i] = p
	}
	return out
}
