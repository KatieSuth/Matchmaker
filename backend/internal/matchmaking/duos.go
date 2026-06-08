package matchmaking

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

const balanceEpsilon = 1e-9

type duoPair struct {
	a, b Player
}

// ApplyDuoLobbyGrouping attempts to place mutual duo partners in the same lobby without
// worsening cross-lobby skill spread compared to the baseline assignment.
func ApplyDuoLobbyGrouping(lobbies []LobbyPlan) []LobbyPlan {
	if len(lobbies) <= 1 {
		return lobbies
	}

	out := cloneLobbies(lobbies)
	baselineSpread := lobbyAverageSpread(out)
	pairs := findMutualDuoPairs(collectRosterPlayers(out))
	sortDuoPairs(pairs)

	for _, pair := range pairs {
		lobbyA, idxA := findPlayerLobby(out, pair.a.UserID)
		lobbyB, idxB := findPlayerLobby(out, pair.b.UserID)
		if lobbyA < 0 || lobbyB < 0 || lobbyA == lobbyB {
			continue
		}

		bestSpread := baselineSpread
		bestSwap := lobbySwap{}
		found := false

		for _, targetLobby := range []int{lobbyA, lobbyB} {
			partnerLobby := lobbyB
			partnerIdx := idxB
			if targetLobby == lobbyB {
				partnerLobby = lobbyA
				partnerIdx = idxA
			}

			for i := range out[targetLobby].Roster {
				trial := cloneLobbies(out)
				trial[targetLobby].Roster[i], trial[partnerLobby].Roster[partnerIdx] =
					trial[partnerLobby].Roster[partnerIdx], trial[targetLobby].Roster[i]

				if !duoInSameLobby(trial, pair.a.UserID, pair.b.UserID) {
					continue
				}

				spread := lobbyAverageSpread(trial)
				if spread > baselineSpread+balanceEpsilon {
					continue
				}
				if !found || spread < bestSpread-balanceEpsilon {
					found = true
					bestSpread = spread
					bestSwap = lobbySwap{
						lobbyA: targetLobby,
						idxA:   i,
						lobbyB: partnerLobby,
						idxB:   partnerIdx,
					}
				}
			}
		}

		if !found {
			continue
		}

		out[bestSwap.lobbyA].Roster[bestSwap.idxA], out[bestSwap.lobbyB].Roster[bestSwap.idxB] =
			out[bestSwap.lobbyB].Roster[bestSwap.idxB], out[bestSwap.lobbyA].Roster[bestSwap.idxA]
	}

	return out
}

type lobbySwap struct {
	lobbyA, idxA, lobbyB, idxB int
}

// applyDuoTeamGrouping attempts to place mutual duo partners on the same team without
// worsening team-average separation compared to the snake-draft baseline.
func applyDuoTeamGrouping(team1, team2 []Player, baselineSep float64) ([]Player, []Player) {
	if len(team1) == 0 && len(team2) == 0 {
		return team1, team2
	}

	roster := append(append([]Player(nil), team1...), team2...)
	pairs := findMutualDuoPairs(roster)
	sortDuoPairs(pairs)

	teamHasDuo := map[int]bool{
		1: pairAlreadyTogether(team1, pairs),
		2: pairAlreadyTogether(team2, pairs),
	}

	for _, pair := range pairs {
		teamA, idxA := findPlayerTeam(team1, team2, pair.a.UserID)
		teamB, idxB := findPlayerTeam(team1, team2, pair.b.UserID)
		if teamA == 0 || teamB == 0 || teamA == teamB {
			continue
		}
		if teamHasDuo[teamA] || teamHasDuo[teamB] {
			continue
		}

		bestSep := baselineSep
		bestSwap := teamSwap{}
		found := false

		for _, targetTeam := range []int{1, 2} {
			if teamHasDuo[targetTeam] {
				continue
			}

			target := team1
			partnerIdx := idxB
			if targetTeam == 2 {
				target = team2
				partnerIdx = idxA
			}

			for i := range target {
				t1, t2 := cloneTeamPair(team1, team2)
				if targetTeam == 1 {
					t1[i], t2[partnerIdx] = t2[partnerIdx], t1[i]
				} else {
					t2[i], t1[partnerIdx] = t1[partnerIdx], t2[i]
				}

				if !duoOnSameTeam(t1, t2, pair.a.UserID, pair.b.UserID) {
					continue
				}

				sep := teamAverageSeparation(t1, t2)
				if sep > baselineSep+balanceEpsilon {
					continue
				}
				if !found || sep < bestSep-balanceEpsilon {
					found = true
					bestSep = sep
					bestSwap = teamSwap{
						targetTeam: targetTeam,
						targetIdx:  i,
						partnerIdx: partnerIdx,
					}
				}
			}
		}

		if !found {
			continue
		}

		if bestSwap.targetTeam == 1 {
			team1[bestSwap.targetIdx], team2[bestSwap.partnerIdx] =
				team2[bestSwap.partnerIdx], team1[bestSwap.targetIdx]
		} else {
			team2[bestSwap.targetIdx], team1[bestSwap.partnerIdx] =
				team1[bestSwap.partnerIdx], team2[bestSwap.targetIdx]
		}
		teamHasDuo[bestSwap.targetTeam] = true
	}

	refreshTeamNumbers(team1, team2)
	return team1, team2
}

type teamSwap struct {
	targetTeam, targetIdx, partnerIdx int
}

// pairAlreadyTogether reports whether any mutual duo pair already shares this team roster.
func pairAlreadyTogether(team []Player, pairs []duoPair) bool {
	ids := make(map[uuid.UUID]struct{}, len(team))
	for _, p := range team {
		ids[p.UserID] = struct{}{}
	}
	for _, pair := range pairs {
		_, aOK := ids[pair.a.UserID]
		_, bOK := ids[pair.b.UserID]
		if aOK && bOK {
			return true
		}
	}
	return false
}

// findMutualDuoPairs returns deduplicated duo pairs whose requests reference each other's Discord names.
func findMutualDuoPairs(players []Player) []duoPair {
	byID := make(map[uuid.UUID]Player, len(players))
	for _, p := range players {
		byID[p.UserID] = p
	}

	seen := make(map[[2]uuid.UUID]struct{})
	var pairs []duoPair

	for _, a := range players {
		if !duoRequestValid(a.DuoRequest) {
			continue
		}
		for _, b := range players {
			if a.UserID == b.UserID {
				continue
			}
			if !isMutualDuoPair(a, b) {
				continue
			}
			key := orderedPairKey(a.UserID, b.UserID)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			pairs = append(pairs, duoPair{a: byID[a.UserID], b: byID[b.UserID]})
		}
	}
	return pairs
}

// isMutualDuoPair is true when both players list each other's Discord name as their duo request.
func isMutualDuoPair(a, b Player) bool {
	if !duoRequestValid(a.DuoRequest) || !duoRequestValid(b.DuoRequest) {
		return false
	}
	return normalizeDuoName(*a.DuoRequest) == normalizeDuoName(b.DiscordName) &&
		normalizeDuoName(*b.DuoRequest) == normalizeDuoName(a.DiscordName)
}

// duoRequestValid is true when a duo request is present and non-empty after trimming.
func duoRequestValid(request *string) bool {
	return request != nil && strings.TrimSpace(*request) != ""
}

// normalizeDuoName trims whitespace and lowercases a Discord name for duo matching.
func normalizeDuoName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// sortDuoPairs orders pairs by earliest registration time, then first player ID, for stable processing.
func sortDuoPairs(pairs []duoPair) {
	for i := 1; i < len(pairs); i++ {
		for j := i; j > 0 && compareDuoPairs(pairs[j], pairs[j-1]) < 0; j-- {
			pairs[j], pairs[j-1] = pairs[j-1], pairs[j]
		}
	}
}

// compareDuoPairs compares two duo pairs for deterministic sort ordering.
func compareDuoPairs(a, b duoPair) int {
	aTime := pairSortTime(a)
	bTime := pairSortTime(b)
	if aTime.Before(bTime) {
		return -1
	}
	if aTime.After(bTime) {
		return 1
	}
	if a.a.UserID.String() < b.a.UserID.String() {
		return -1
	}
	if a.a.UserID.String() > b.a.UserID.String() {
		return 1
	}
	return 0
}

// pairSortTime returns the earlier CreatedAt timestamp from a duo pair.
func pairSortTime(pair duoPair) time.Time {
	t := pair.a.CreatedAt
	if pair.b.CreatedAt.Before(t) {
		t = pair.b.CreatedAt
	}
	return t
}

// orderedPairKey builds a stable deduplication key for two user IDs regardless of order.
func orderedPairKey(a, b uuid.UUID) [2]uuid.UUID {
	if a.String() < b.String() {
		return [2]uuid.UUID{a, b}
	}
	return [2]uuid.UUID{b, a}
}

// lobbyAverageSpread returns max(lobby avg rank) - min(lobby avg rank) across non-empty lobbies.
func lobbyAverageSpread(lobbies []LobbyPlan) float64 {
	var avgs []float64
	for _, lobby := range lobbies {
		if len(lobby.Roster) == 0 {
			continue
		}
		avgs = append(avgs, averageRank(lobby.Roster))
	}
	if len(avgs) < 2 {
		return 0
	}
	minAvg, maxAvg := avgs[0], avgs[0]
	for _, avg := range avgs[1:] {
		if avg < minAvg {
			minAvg = avg
		}
		if avg > maxAvg {
			maxAvg = avg
		}
	}
	return maxAvg - minAvg
}

// averageRank returns the mean AvgRank for a player slice, or zero when empty.
func averageRank(players []Player) float64 {
	if len(players) == 0 {
		return 0
	}
	var sum float64
	for _, p := range players {
		sum += p.AvgRank
	}
	return sum / float64(len(players))
}

// teamAverageSeparation returns the absolute difference between two teams' mean AvgRank values.
func teamAverageSeparation(team1, team2 []Player) float64 {
	avg1 := averageRank(team1)
	avg2 := averageRank(team2)
	if len(team1) == 0 || len(team2) == 0 {
		return 0
	}
	return math.Abs(avg1 - avg2)
}

// collectRosterPlayers flattens all lobby rosters into a single player slice.
func collectRosterPlayers(lobbies []LobbyPlan) []Player {
	var players []Player
	for _, lobby := range lobbies {
		players = append(players, lobby.Roster...)
	}
	return players
}

// findPlayerLobby locates a rostered player; returns (-1, -1) when not found.
func findPlayerLobby(lobbies []LobbyPlan, userID uuid.UUID) (lobbyIdx, rosterIdx int) {
	for i, lobby := range lobbies {
		for j, p := range lobby.Roster {
			if p.UserID == userID {
				return i, j
			}
		}
	}
	return -1, -1
}

// findPlayerTeam locates a player on team 1 or 2; returns (0, -1) when not found.
func findPlayerTeam(team1, team2 []Player, userID uuid.UUID) (teamNum, idx int) {
	for i, p := range team1 {
		if p.UserID == userID {
			return 1, i
		}
	}
	for i, p := range team2 {
		if p.UserID == userID {
			return 2, i
		}
	}
	return 0, -1
}

// cloneLobbies returns a deep copy of lobby plans for non-destructive swap evaluation.
func cloneLobbies(lobbies []LobbyPlan) []LobbyPlan {
	out := make([]LobbyPlan, len(lobbies))
	for i, lobby := range lobbies {
		out[i] = LobbyPlan{
			Roster:          append([]Player(nil), lobby.Roster...),
			Subs:            append([]Player(nil), lobby.Subs...),
			HostID:          lobby.HostID,
			FairnessWarning: lobby.FairnessWarning,
		}
	}
	return out
}

// cloneTeamPair returns deep copies of both team rosters for swap evaluation.
func cloneTeamPair(team1, team2 []Player) ([]Player, []Player) {
	return append([]Player(nil), team1...), append([]Player(nil), team2...)
}

// duoInSameLobby reports whether both players are rostered in the same lobby.
func duoInSameLobby(lobbies []LobbyPlan, userA, userB uuid.UUID) bool {
	lobbyA, _ := findPlayerLobby(lobbies, userA)
	lobbyB, _ := findPlayerLobby(lobbies, userB)
	return lobbyA >= 0 && lobbyA == lobbyB
}

// duoOnSameTeam reports whether both players are assigned to the same team number.
func duoOnSameTeam(team1, team2 []Player, userA, userB uuid.UUID) bool {
	teamA, _ := findPlayerTeam(team1, team2, userA)
	teamB, _ := findPlayerTeam(team1, team2, userB)
	return teamA != 0 && teamA == teamB
}

// refreshTeamNumbers reassigns team_number after roster swaps so values match slice placement.
func refreshTeamNumbers(team1, team2 []Player) {
	for i := range team1 {
		team1[i].TeamNumber = intPtr(1)
	}
	for i := range team2 {
		team2[i].TeamNumber = intPtr(2)
	}
}
