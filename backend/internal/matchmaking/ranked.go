package matchmaking

// AssignRanked packs contiguous rank bands into lobbies (ascending rank order).
func AssignRanked(players []Player, lobbyCount, slotsPerLobby int) []LobbyPlan {
	lobbies := make([]LobbyPlan, lobbyCount)
	for i := range lobbies {
		lobbies[i].Roster = make([]Player, 0, slotsPerLobby)
	}

	needed := lobbyCount * slotsPerLobby
	pool := selectRankedRosterPool(players, needed)

	idx := 0
	for lobbyIdx := 0; lobbyIdx < lobbyCount && idx < len(pool); lobbyIdx++ {
		for slot := 0; slot < slotsPerLobby && idx < len(pool); slot++ {
			lobbies[lobbyIdx].Roster = append(lobbies[lobbyIdx].Roster, pool[idx])
			idx++
		}
	}
	return lobbies
}

// sortPlayersByRankAsc orders players lowest skill first, using availability as a tie-break.
func sortPlayersByRankAsc(players []Player) {
	for i := 1; i < len(players); i++ {
		for j := i; j > 0 && CompareByRankThenAvailability(players[j-1], players[j]) < 0; j-- {
			players[j], players[j-1] = players[j-1], players[j]
		}
	}
}
