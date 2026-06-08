package matchmaking

// Test exports for black-box coverage of duo helpers.

type DuoPairForTest struct {
	A, B Player
}

var (
	LobbyAverageSpreadForTest      = lobbyAverageSpread
	TeamAverageSeparationForTest   = teamAverageSeparation
	SplitRosterByTeamNumberForTest = splitRosterByTeamNumber
	ApplyDuoTeamGroupingForTest    = applyDuoTeamGrouping
	SnakeDraftIntoTeamsForTest     = snakeDraftIntoTeams
	FindMutualDuoPairsForTest    = func(players []Player) []DuoPairForTest {
		pairs := findMutualDuoPairs(players)
		out := make([]DuoPairForTest, len(pairs))
		for i, pair := range pairs {
			out[i] = DuoPairForTest{A: pair.a, B: pair.b}
		}
		return out
	}
	SortMutualDuoPairsForTest = func(players []Player) []DuoPairForTest {
		pairs := findMutualDuoPairs(players)
		sortDuoPairs(pairs)
		out := make([]DuoPairForTest, len(pairs))
		for i, pair := range pairs {
			out[i] = DuoPairForTest{A: pair.a, B: pair.b}
		}
		return out
	}
	CompareDuoPairsForTest = func(a, b DuoPairForTest) int {
		return compareDuoPairs(duoPair{a: a.A, b: a.B}, duoPair{a: b.A, b: b.B})
	}
	FindPlayerLobbyForTest = findPlayerLobby
	FindPlayerTeamForTest  = findPlayerTeam
	SwapRosterPlayerForTest = swapRosterPlayer
)
