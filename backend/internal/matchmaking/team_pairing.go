package matchmaking

import (
	"math"
	"sort"
)

// lobbyTeam is one already-split team carried between lobbies by pairTeamsAcrossLobbies.
type lobbyTeam struct {
	players []Player
	avg     float64
}

// pairTeamsAcrossLobbies re-pairs balanced-mode teams across lobbies so each lobby's two teams
// have the closest averages: teams are sorted by average and neighbors share a lobby, which is
// the optimal pairing for minimizing team-average gaps. Teams are moved intact. The new pairing
// is kept only if it lowers the worst team-average gap and leaves no more lobbies unfair; a
// lobby-united duo on opposite teams may be split across lobbies, since balance wins.
// Single-lobby games and lobbies without two seated teams are left unchanged.
func pairTeamsAcrossLobbies(lobbies []LobbyPlan, settings Settings, tierCount int) []LobbyPlan {
	if len(lobbies) < 2 {
		return lobbies
	}
	teams := make([]lobbyTeam, 0, 2*len(lobbies))
	for _, lobby := range lobbies {
		team1, team2 := splitRosterByTeamNumber(lobby.Roster)
		if len(team1) == 0 || len(team2) == 0 || len(lobby.Subs) > 0 {
			return lobbies
		}
		teams = append(teams, lobbyTeam{players: team1, avg: averageRank(team1)}, lobbyTeam{players: team2, avg: averageRank(team2)})
	}
	sort.SliceStable(teams, func(i, j int) bool { return teams[i].avg > teams[j].avg })

	repaired := cloneLobbies(lobbies)
	for i := range repaired {
		repaired[i].Roster = append(withTeamNumber(teams[2*i].players, 1), withTeamNumber(teams[2*i+1].players, 2)...)
	}

	if !(worstTeamGap(repaired) < worstTeamGap(lobbies)-balanceEpsilon) {
		return lobbies
	}
	if countUnfairLobbies(repaired, settings, tierCount) > countUnfairLobbies(lobbies, settings, tierCount) {
		return lobbies
	}
	return repaired
}

// withTeamNumber copies players onto team n.
func withTeamNumber(players []Player, n int) []Player {
	out := make([]Player, len(players))
	for i, p := range players {
		p.TeamNumber = intPtr(n)
		out[i] = p
	}
	return out
}

// worstTeamGap is the largest team-average separation across lobbies.
func worstTeamGap(lobbies []LobbyPlan) float64 {
	worst := 0.0
	for _, lobby := range lobbies {
		worst = math.Max(worst, lobbyTeamSeparation(lobby))
	}
	return worst
}

// countUnfairLobbies is how many lobbies fail either fairness check.
func countUnfairLobbies(lobbies []LobbyPlan, settings Settings, tierCount int) int {
	n := 0
	for _, lobby := range lobbies {
		if IsLobbyUnfair(lobby, settings, tierCount) {
			n++
		}
	}
	return n
}
