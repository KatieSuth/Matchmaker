package matchmaking

import "github.com/google/uuid"

// Test exports for black-box coverage of unexported matchmaking helpers.

type DuoPairForTest struct {
	A, B Player
}

var (
	LobbyAverageSpreadForTest      = lobbyAverageSpread
	TeamAverageSeparationForTest   = teamAverageSeparation
	SplitRosterByTeamNumberForTest = splitRosterByTeamNumber
	ApplyDuoTeamGroupingForTest    = applyDuoTeamGrouping
	SnakeDraftIntoTeamsForTest     = snakeDraftIntoTeams
	WindowDraftIntoTeamsForTest    = windowDraftIntoTeams
	SplitIntoTeamsWindowedForTest  = splitIntoTeamsWindowed
	ClampedRankOrderForTest        = clampedRankOrder
	PickClosestToMidpointForTest   = pickClosestToMidpoint
	FullestWindowIndexForTest      = fullestWindowIndex

	FindMutualDuoPairsForTest = func(players []Player) []DuoPairForTest {
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
	FindPlayerLobbyForTest  = findPlayerLobby
	FindPlayerTeamForTest   = findPlayerTeam
	SwapRosterPlayerForTest = swapRosterPlayer

	RepairUnfairWindowPairsForTest        = repairUnfairWindowPairs
	IndexCombinationsForTest              = indexCombinations
	LeftoverCanJoinRosterForTest          = leftoverCanJoinRoster
	ReplaceRosterPlayerKeepingTeamForTest = replaceRosterPlayerKeepingTeam
	RankSpreadForTest                     = rankSpread
	CapWindowSearchPoolForTest            = capWindowSearchPool
	OutlierGapForTest                     = outlierGap
	ShouldDropReservedSubLobbyForTest     = shouldDropReservedSubLobby
	AnyLobbyUnfairForTest                 = anyLobbyUnfair
	AssignBalancedSnakeForTest            = assignBalancedSnake
	SnakeLobbyIndexForTest                = snakeLobbyIndex
	AppendToWeakerTeamForTest             = appendToWeakerTeam
)

func AdjacentFringeSwapOKForTest(rosteredRank, leftoverRank float64, teamSize, tierCount int) bool {
	windows := buildRankWindows(tierCount, teamSize)
	ro := clampedRankOrder(rosteredRank, tierCount)
	lo := clampedRankOrder(leftoverRank, tierCount)
	rw := windowIndexForOrder(ro, windows)
	lw := windowIndexForOrder(lo, windows)
	return adjacentFringeSwapOK(rw, lw, ro, lo, windows)
}

func RepairSwapAllowedForTest(rosteredRank, leftoverRank float64, teamSize, tierCount int, sameWindowOnly bool) bool {
	windows := buildRankWindows(tierCount, teamSize)
	return repairSwapAllowed(
		Player{AvgRank: rosteredRank},
		Player{AvgRank: leftoverRank},
		windows,
		tierCount,
		sameWindowOnly,
	)
}

func WindowSpreadWorsensForTest(roster []Player, outID uuid.UUID, in Player, teamSize, tierCount int) bool {
	return windowSpreadWorsens(roster, outID, in, buildRankWindows(tierCount, teamSize), tierCount)
}

func RepairTrialAcceptedForTest(before, after LobbyPlan, settings Settings, tierCount int) bool {
	return repairTrialAccepted(
		captureRepairState(before, settings, tierCount),
		captureRepairState(after, settings, tierCount),
	)
}

func InPlaceTrialNotBetterForTest(
	optimizeSep bool,
	bestSep, bestOutlier, trialSep, trialOutlier, delta, bestDelta float64,
	rostered, leftover, bestOut, bestIn Player,
) bool {
	return inPlaceTrialNotBetter(
		optimizeSep,
		repairState{sep: bestSep, outlier: bestOutlier},
		repairState{sep: trialSep, outlier: trialOutlier},
		delta,
		bestDelta,
		rostered,
		leftover,
		bestOut,
		bestIn,
	)
}

func TryRepairUnfairLobbyWindowComboForTest(
	lobby LobbyPlan,
	leftovers []Player,
	unrosteredCanSubs, reserved int,
	cfg Config,
	settings Settings,
) (LobbyPlan, bool) {
	windows := buildRankWindows(cfg.TierCount, cfg.TeamSize)
	ok := tryRepairUnfairLobbyWindowCombo(&lobby, leftovers, unrosteredCanSubs, reserved, cfg, windows, settings)
	return lobby, ok
}

func WindowIndexForOrderForTest(order int, windows []RankWindowForTest) int {
	converted := make([]rankWindow, len(windows))
	for i, w := range windows {
		converted[i] = rankWindow{minOrder: w.MinOrder, maxOrder: w.MaxOrder, midpoint: w.Midpoint}
	}
	return windowIndexForOrder(order, converted)
}

func SamePlayerSetForTest(players []Player, ids []uuid.UUID) bool {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return samePlayerSet(players, set)
}

type WindowLeftoverForTest struct {
	Player Player
	Window int
	Order  int
}

func leftoverFromTest(items []WindowLeftoverForTest) []windowLeftover {
	out := make([]windowLeftover, len(items))
	for i, item := range items {
		out[i] = windowLeftover{player: item.Player, window: item.Window, order: item.Order}
	}
	return out
}

func BestLeftoverPairFromTest(items []WindowLeftoverForTest) (int, int) {
	return bestLeftoverPair(leftoverFromTest(items))
}

func RemoveLeftoverPairFromTest(items []WindowLeftoverForTest, i, j int) []WindowLeftoverForTest {
	removed := removeLeftoverPair(leftoverFromTest(items), i, j)
	out := make([]WindowLeftoverForTest, len(removed))
	for k, item := range removed {
		out[k] = WindowLeftoverForTest{Player: item.player, Window: item.window, Order: item.order}
	}
	return out
}

func BestNonSubReplacementForTest(allPlayers []Player, lobbies []LobbyPlan, vacated Player, teamSize, tierCount int) (Player, bool) {
	return bestNonSubReplacement(allPlayers, lobbies, vacated, buildRankWindows(tierCount, teamSize), tierCount)
}

func PickCloserToTargetForTest(players []Player, target float64) Player {
	return pickCloserToTarget(players, target)
}

type RankWindowForTest struct {
	MinOrder int
	MaxOrder int
	Midpoint float64
}

func BuildRankWindowsForTest(tierCount, teamSize int) []RankWindowForTest {
	windows := buildRankWindows(tierCount, teamSize)
	out := make([]RankWindowForTest, len(windows))
	for i, w := range windows {
		out[i] = RankWindowForTest{MinOrder: w.minOrder, MaxOrder: w.maxOrder, Midpoint: w.midpoint}
	}
	return out
}
