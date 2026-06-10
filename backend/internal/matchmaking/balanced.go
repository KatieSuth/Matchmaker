package matchmaking

// AssignBalanced distributes roster players across lobbies via snake draft by skill.
func AssignBalanced(players []Player, lobbyCount, slotsPerLobby int) []LobbyPlan {
	lobbies := make([]LobbyPlan, lobbyCount)
	for i := range lobbies {
		lobbies[i].Roster = make([]Player, 0, slotsPerLobby)
	}

	needed := lobbyCount * slotsPerLobby
	pool := selectBalancedRosterPool(players, needed)
	sortPlayersByRankAsc(pool)

	for i, p := range pool {
		lobbyIdx := snakeLobbyIndex(i, lobbyCount)
		lobbies[lobbyIdx].Roster = append(lobbies[lobbyIdx].Roster, p)
	}
	return lobbies
}

// snakeLobbyIndex maps a draft pick number to a lobby index using alternating direction each round.
func snakeLobbyIndex(i, lobbyCount int) int {
	round := i / lobbyCount
	pos := i % lobbyCount
	if round%2 == 1 {
		pos = lobbyCount - 1 - pos
	}
	return pos
}
