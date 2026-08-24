package matchmaking

import (
	"math"

	"github.com/google/uuid"
)

// Cap the per-window search so C(n, k) stays small. Typical 2v2/5v5 pools are far below this.
const maxWindowSearchPool = 12

// Adjacent leftover ranks this close to a shared window edge may 1-for-1 with
// a rostered player on the other side. Wide enough for Gold 1 ↔ Plat 1 on a
// 25-tier 5v5 ladder; too tight to dump a Radiant into a Gold hole.
const repairFringeRanks = 2

// repairUnfairWindowPairs pulls unplaced leftovers into an unfair lobby when
// that strictly improves the check that is failing (team-average separation
// and/or outlier gap). Ranked mode never runs this: contiguous rank packing is
// the product.
func repairUnfairWindowPairs(lobbies []LobbyPlan, allPlayers []Player, cfg Config, settings Settings) []LobbyPlan {
	if len(lobbies) == 0 || cfg.SortLogic == "ranked" {
		return lobbies
	}
	windows := buildRankWindows(cfg.TierCount, cfg.TeamSize)
	if len(windows) == 0 {
		return lobbies
	}

	maxPasses := len(windows) + 2
	if maxPasses > 8 {
		maxPasses = 8
	}

	for i := range lobbies {
		// One improving swap per pass; stop when the lobby is fair or nothing helps.
		for pass := 0; pass < maxPasses; pass++ {
			if !lobbyNeedsFairnessRepair(lobbies[i], settings, cfg.TierCount) {
				break
			}
			if !tryRepairUnfairLobbyWindow(&lobbies[i], lobbies, allPlayers, cfg, windows, settings) {
				break
			}
		}
	}
	return lobbies
}

// lobbyNeedsFairnessRepair is true when either fairness check is already failing.
func lobbyNeedsFairnessRepair(lobby LobbyPlan, settings Settings, tierCount int) bool {
	return IsLobbyUnfair(lobby, settings, tierCount)
}

// repairState is the pair of fairness metrics a trial is scored against.
type repairState struct {
	sep, outlier     float64
	sepOver, outOver bool
}

// captureRepairState snapshots both fairness metrics so a trial can be compared
// against the live lobby without re-reading thresholds in every helper.
func captureRepairState(lobby LobbyPlan, settings Settings, tierCount int) repairState {
	th := ScaledFairnessThresholds(settings, tierCount)
	return repairState{
		sep:     lobbyTeamSeparation(lobby),
		outlier: outlierGap(lobby.Roster),
		sepOver: teamSeparationExceeds(lobby, th.TeamSeparation),
		outOver: outlierExceeds(lobby.Roster, th.OutlierGap),
	}
}

// repairTrialAccepted is true when the trial strictly improves the failing
// check and does not newly trip the other one.
func repairTrialAccepted(before, after repairState) bool {
	if before.sepOver {
		if after.sep >= before.sep-balanceEpsilon {
			return false
		}
		// Shrinking team sep must not create a new outlier warning.
		if !before.outOver && after.outOver {
			return false
		}
		return true
	}
	if before.outOver {
		if after.outlier >= before.outlier-balanceEpsilon {
			return false
		}
		// Shrinking the outlier must not create a team-sep warning.
		if after.sepOver {
			return false
		}
		return true
	}
	return false
}

// repairPrimary is the metric currently being minimized: team sep if that check
// failed, otherwise outlier gap.
func repairPrimary(state repairState, optimizeSep bool) float64 {
	if optimizeSep {
		return state.sep
	}
	return state.outlier
}

// tryRepairUnfairLobbyWindow searches one improving leftover swap. Leftovers
// are unplaced players only — other lobbies are never raided. In-place 1-for-1
// (same window, then adjacent fringe) runs before a full-window re-split so a
// near-fringe leftover can sit in the vacated seat without reshuffling the lobby.
func tryRepairUnfairLobbyWindow(lobby *LobbyPlan, lobbies []LobbyPlan, allPlayers []Player, cfg Config, windows []rankWindow, settings Settings) bool {
	leftovers := unplacedPlayers(allPlayers, lobbies)
	unrosteredCanSubs := countUnrosteredCanSubs(allPlayers, lobbies)
	reserved := reservedCanSubs(cfg)

	if tryInPlaceLeftoverSwap(lobby, leftovers, unrosteredCanSubs, reserved, windows, cfg, settings, true) {
		return true
	}
	if tryInPlaceLeftoverSwap(lobby, leftovers, unrosteredCanSubs, reserved, windows, cfg, settings, false) {
		return true
	}
	return tryRepairUnfairLobbyWindowCombo(lobby, leftovers, unrosteredCanSubs, reserved, cfg, windows, settings)
}

// leftoverCanJoinRoster is false when taking this can-sub would leave too few
// unrostered volunteers for n × sub_min.
func leftoverCanJoinRoster(p Player, unrosteredCanSubs, reserved int) bool {
	if !p.CanSubstitute {
		return true
	}
	return unrosteredCanSubs-1 >= reserved
}

// tryInPlaceLeftoverSwap replaces one rostered player with one leftover and
// keeps every other team_number. sameWindowOnly considers the leftover's band;
// otherwise only adjacent-window pairs within repairFringeRanks of the shared edge.
func tryInPlaceLeftoverSwap(
	lobby *LobbyPlan,
	leftovers []Player,
	unrosteredCanSubs, reserved int,
	windows []rankWindow,
	cfg Config,
	settings Settings,
	sameWindowOnly bool,
) bool {
	before := captureRepairState(*lobby, settings, cfg.TierCount)
	// Both checks failing: even the sides first; the outlier often follows.
	optimizeSep := before.sepOver
	bestState := before
	bestDelta := math.Inf(1)
	found := false
	var bestRoster []Player
	var bestOut, bestIn Player

	for _, leftover := range leftovers {
		if !leftoverCanJoinRoster(leftover, unrosteredCanSubs, reserved) {
			continue
		}
		for _, rostered := range lobby.Roster {
			// Subs have no seat; swapping them would not change team averages.
			if rostered.TeamNumber == nil {
				continue
			}
			if !repairSwapAllowed(rostered, leftover, windows, cfg.TierCount, sameWindowOnly) {
				continue
			}
			// Same-window only: a leftover Radiant must not replace an Ascendant
			// to paper over Plat vs Bronze with a wider band.
			if sameWindowOnly && windowSpreadWorsens(lobby.Roster, rostered.UserID, leftover, windows, cfg.TierCount) {
				continue
			}

			trial := replaceRosterPlayerKeepingTeam(lobby.Roster, rostered.UserID, leftover)
			trialState := captureRepairState(LobbyPlan{Roster: trial}, settings, cfg.TierCount)
			if !repairTrialAccepted(before, trialState) {
				continue
			}
			delta := math.Abs(rostered.AvgRank - leftover.AvgRank)
			if found && inPlaceTrialNotBetter(optimizeSep, bestState, trialState, delta, bestDelta, rostered, leftover, bestOut, bestIn) {
				continue
			}

			bestState = trialState
			bestDelta = delta
			bestRoster = trial
			bestOut = rostered
			bestIn = leftover
			found = true
		}
	}

	if !found {
		return false
	}
	lobby.Roster = bestRoster
	return true
}

// inPlaceTrialNotBetter is true when trial should not replace the current best
// 1-for-1. Primary metric first, then closer ranks, then preferInPlaceSwap.
func inPlaceTrialNotBetter(
	optimizeSep bool,
	best, trial repairState,
	delta, bestDelta float64,
	rostered, leftover, bestOut, bestIn Player,
) bool {
	bestPrimary := repairPrimary(best, optimizeSep)
	trialPrimary := repairPrimary(trial, optimizeSep)
	if trialPrimary > bestPrimary+balanceEpsilon {
		return true
	}
	if trialPrimary >= bestPrimary-balanceEpsilon {
		if delta > bestDelta+balanceEpsilon {
			return true
		}
		if delta >= bestDelta-balanceEpsilon && !preferInPlaceSwap(rostered, leftover, bestOut, bestIn) {
			return true
		}
	}
	return false
}

// preferInPlaceSwap breaks a sep+delta tie so the swap that actually fills the
// weak seat wins over a numerically identical high-for-high exchange.
func preferInPlaceSwap(out, in, bestOut, bestIn Player) bool {
	// On a sep+delta tie, replace the lower-ranked seat so a high-band leftover
	// cannot mask a low-window hole with the same numeric improvement.
	if cmp := CompareByRankThenAvailability(out, bestOut); cmp != 0 {
		return cmp > 0
	}
	return CompareByRankThenAvailability(in, bestIn) < 0
}

// windowSpreadWorsens is true when swapping leftover in for outID would widen
// that window's min–max rank span. Repair may not paper over a lopsided pair
// by stretching the band.
func windowSpreadWorsens(roster []Player, outID uuid.UUID, in Player, windows []rankWindow, tierCount int) bool {
	w := windowIndexForOrder(clampedRankOrder(in.AvgRank, tierCount), windows)
	var before, after []float64
	for _, p := range roster {
		if windowIndexForOrder(clampedRankOrder(p.AvgRank, tierCount), windows) != w {
			continue
		}
		before = append(before, p.AvgRank)
		if p.UserID == outID {
			after = append(after, in.AvgRank)
			continue
		}
		after = append(after, p.AvgRank)
	}
	return rankSpread(after) > rankSpread(before)+balanceEpsilon
}

// rankSpread is max(ranks) − min(ranks). A single rank has no spread.
func rankSpread(ranks []float64) float64 {
	if len(ranks) < 2 {
		return 0
	}
	minR, maxR := ranks[0], ranks[0]
	for _, r := range ranks[1:] {
		if r < minR {
			minR = r
		}
		if r > maxR {
			maxR = r
		}
	}
	return maxR - minR
}

// repairSwapAllowed is the window gate for a 1-for-1. Same-window mode requires
// both players in one band; fringe mode requires neighboring bands near the edge.
func repairSwapAllowed(rostered, leftover Player, windows []rankWindow, tierCount int, sameWindowOnly bool) bool {
	ro := clampedRankOrder(rostered.AvgRank, tierCount)
	lo := clampedRankOrder(leftover.AvgRank, tierCount)
	rw := windowIndexForOrder(ro, windows)
	lw := windowIndexForOrder(lo, windows)
	if sameWindowOnly {
		return rw == lw
	}
	return adjacentFringeSwapOK(rw, lw, ro, lo, windows)
}

// adjacentFringeSwapOK allows a leftover in a neighboring window only when both
// ranks sit within repairFringeRanks of the shared edge (Gold 1 ↔ Plat 1, not
// Gold 2 vs a mid-Plat, and never Radiant into Gold).
func adjacentFringeSwapOK(rw, lw, ro, lo int, windows []rankWindow) bool {
	if len(windows) == 0 || absInt(rw-lw) != 1 {
		return false
	}
	lower, higher := rw, lw
	if lw < rw {
		lower, higher = lw, rw
	}
	edgeLow := windows[lower].maxOrder
	edgeHigh := windows[higher].minOrder
	return fringeDistance(ro, rw, lower, edgeLow, edgeHigh) <= repairFringeRanks &&
		fringeDistance(lo, lw, lower, edgeLow, edgeHigh) <= repairFringeRanks
}

// fringeDistance is how many ladder steps a rank sits from the shared window
// edge, measured inward from that player's own band.
func fringeDistance(order, window, lower, edgeLow, edgeHigh int) int {
	if window == lower {
		return edgeLow - order
	}
	return order - edgeHigh
}

// replaceRosterPlayerKeepingTeam copies roster with leftover sitting in outID's
// team_number so the rest of the lobby is not re-dealt.
func replaceRosterPlayerKeepingTeam(roster []Player, outID uuid.UUID, in Player) []Player {
	out := make([]Player, len(roster))
	for i, p := range roster {
		if p.UserID != outID {
			out[i] = p
			continue
		}
		if p.TeamNumber == nil {
			out[i] = p
			continue
		}
		in.TeamNumber = intPtr(*p.TeamNumber)
		out[i] = in
	}
	return out
}

// tryRepairUnfairLobbyWindowCombo retries the k rostered seats in one window
// as combinations with unplaced players from that window, then re-splits the
// whole lobby. Used only when no 1-for-1 improves fairness. Ties keep more of
// the current window roster so a re-deal does not churn seats for the same score.
func tryRepairUnfairLobbyWindowCombo(
	lobby *LobbyPlan,
	leftovers []Player,
	unrosteredCanSubs, reserved int,
	cfg Config,
	windows []rankWindow,
	settings Settings,
) bool {
	before := captureRepairState(*lobby, settings, cfg.TierCount)
	// Both checks failing: even the sides first; the outlier often follows.
	optimizeSep := before.sepOver
	bestState := before
	found := false
	bestOverlap := 0
	var bestRoster []Player

	for w := range windows {
		rosteredInWindow := playersInWindow(lobby.Roster, w, windows, cfg.TierCount)
		k := len(rosteredInWindow)
		if k == 0 {
			continue
		}
		windowLeftovers := playersInWindow(leftovers, w, windows, cfg.TierCount)
		if len(windowLeftovers) == 0 {
			continue
		}

		// Always keep the current k occupants; leftovers are trimmed to the
		// nearest midpoint ranks if C(n, k) would explode.
		pool := capWindowSearchPool(rosteredInWindow, windowLeftovers, windows[w].midpoint)

		currentIDs := playerIDSet(rosteredInWindow)
		combos := indexCombinations(len(pool), k)
		for _, combo := range combos {
			chosen := chosenPlayers(pool, combo)
			// Same k people: a re-split of the current window is not a leftover repair.
			if samePlayerSet(chosen, currentIDs) {
				continue
			}
			// Same rule as in-place: a leftover Radiant must not stretch this
			// band to paper over Plat vs Bronze.
			if rankSpread(playerRanks(chosen)) > rankSpread(playerRanks(rosteredInWindow))+balanceEpsilon {
				continue
			}
			added := leftoverCanSubsAdded(chosen, currentIDs)
			// Promoting leftover can-subs must still leave n × sub_min unrostered.
			if unrosteredCanSubs-added < reserved {
				continue
			}

			// Other windows keep their people; this window is swapped in as a
			// block, then the whole lobby is re-dealt so team_numbers stay valid.
			trialRoster := replaceWindowOnRoster(lobby.Roster, w, chosen, windows, cfg.TierCount)
			team1, team2 := splitIntoTeamsWindowed(trialRoster, cfg.TeamSize, cfg.TierCount)
			trialState := captureRepairState(LobbyPlan{Roster: append(team1, team2...)}, settings, cfg.TierCount)
			if !repairTrialAccepted(before, trialState) {
				continue
			}
			overlap := overlapCount(chosen, currentIDs)
			if found {
				bestPrimary := repairPrimary(bestState, optimizeSep)
				trialPrimary := repairPrimary(trialState, optimizeSep)
				if trialPrimary > bestPrimary+balanceEpsilon {
					continue
				}
				// Equal primary: keep the combo that retains more of the current
				// window roster so a re-split is not a full reshuffle for a tie.
				if trialPrimary >= bestPrimary-balanceEpsilon && overlap <= bestOverlap {
					continue
				}
			}

			bestState = trialState
			bestOverlap = overlap
			bestRoster = append(team1, team2...)
			found = true
		}
	}

	if !found {
		return false
	}
	lobby.Roster = bestRoster
	return true
}

// reservedCanSubs is n × sub_min, the volunteer floor that leftover repair
// must not raid. A single lobby has no per-lobby sub quota, so the floor is 0.
func reservedCanSubs(cfg Config) int {
	if cfg.LobbyCount < 2 || cfg.SubMin <= 0 {
		return 0
	}
	return cfg.LobbyCount * cfg.SubMin
}

// unplacedPlayers are registrations with no roster or sub seat. rosteredUserIDs
// includes subs, so a leftover cannot steal a player already assigned as a sub.
func unplacedPlayers(allPlayers []Player, lobbies []LobbyPlan) []Player {
	assigned := rosteredUserIDs(lobbies)
	var out []Player
	for _, p := range allPlayers {
		if assigned[p.UserID] {
			continue
		}
		out = append(out, p)
	}
	return out
}

// countUnrosteredCanSubs is how many sub-eligible players are still unplaced.
// Repair consults this before promoting a leftover can-sub onto a roster.
func countUnrosteredCanSubs(allPlayers []Player, lobbies []LobbyPlan) int {
	assigned := rosteredUserIDs(lobbies)
	n := 0
	for _, p := range allPlayers {
		if p.CanSubstitute && !assigned[p.UserID] {
			n++
		}
	}
	return n
}

// playersInWindow filters to the rank band at windowIdx.
func playersInWindow(players []Player, windowIdx int, windows []rankWindow, tierCount int) []Player {
	var out []Player
	for _, p := range players {
		if windowIndexForOrder(clampedRankOrder(p.AvgRank, tierCount), windows) == windowIdx {
			out = append(out, p)
		}
	}
	return out
}

// capWindowSearchPool always keeps the full current window roster and, if the
// leftover list would blow C(n, k), keeps leftovers closest to the window midpoint.
func capWindowSearchPool(rostered, leftovers []Player, midpoint float64) []Player {
	if len(rostered)+len(leftovers) <= maxWindowSearchPool {
		out := make([]Player, 0, len(rostered)+len(leftovers))
		out = append(out, rostered...)
		return append(out, leftovers...)
	}
	if len(rostered) >= maxWindowSearchPool {
		return append([]Player(nil), rostered...)
	}
	want := maxWindowSearchPool - len(rostered)
	picked, _ := pickClosestToMidpoint(leftovers, want, midpoint)
	out := make([]Player, 0, len(rostered)+len(picked))
	out = append(out, rostered...)
	return append(out, picked...)
}

// chosenPlayers materializes a combination. TeamNumber is cleared so the
// following windowed re-split assigns seats from a blank slate.
func chosenPlayers(pool []Player, combo []int) []Player {
	chosen := make([]Player, len(combo))
	for i, idx := range combo {
		p := pool[idx]
		p.TeamNumber = nil
		chosen[i] = p
	}
	return chosen
}

// playerRanks extracts AvgRank for spread comparisons.
func playerRanks(players []Player) []float64 {
	ranks := make([]float64, len(players))
	for i, p := range players {
		ranks[i] = p.AvgRank
	}
	return ranks
}

// leftoverCanSubsAdded counts sub-eligible players in chosen who are not
// already on this window's roster — the drain against the reserved can-sub floor.
func leftoverCanSubsAdded(chosen []Player, alreadyRostered map[uuid.UUID]struct{}) int {
	n := 0
	for _, p := range chosen {
		if !p.CanSubstitute {
			continue
		}
		if _, ok := alreadyRostered[p.UserID]; ok {
			continue
		}
		n++
	}
	return n
}

// replaceWindowOnRoster drops the current occupants of windowIdx and appends
// chosen in their place. Other windows keep their seats until the re-split.
func replaceWindowOnRoster(roster []Player, windowIdx int, chosen []Player, windows []rankWindow, tierCount int) []Player {
	out := make([]Player, 0, len(roster))
	for _, p := range roster {
		if windowIndexForOrder(clampedRankOrder(p.AvgRank, tierCount), windows) == windowIdx {
			continue
		}
		out = append(out, p)
	}
	return append(out, chosen...)
}

// playerIDSet is the identity of a window roster, used to skip no-op combos
// and to score overlap on a tied fairness improvement.
func playerIDSet(players []Player) map[uuid.UUID]struct{} {
	ids := make(map[uuid.UUID]struct{}, len(players))
	for _, p := range players {
		ids[p.UserID] = struct{}{}
	}
	return ids
}

// samePlayerSet is true when players is exactly the ID set — the combo that
// would re-split the window without changing who is in it.
func samePlayerSet(players []Player, ids map[uuid.UUID]struct{}) bool {
	if len(players) != len(ids) {
		return false
	}
	for _, p := range players {
		if _, ok := ids[p.UserID]; !ok {
			return false
		}
	}
	return true
}

// overlapCount is how many chosen players already occupy this window. Combo
// ties prefer a higher count so the re-deal keeps more of the current roster.
func overlapCount(players []Player, ids map[uuid.UUID]struct{}) int {
	n := 0
	for _, p := range players {
		if _, ok := ids[p.UserID]; ok {
			n++
		}
	}
	return n
}

// lobbyTeamSeparation is |avg(team 1) − avg(team 2)|. Subs (nil team_number)
// are excluded by the split, matching the host-facing team-sep check.
func lobbyTeamSeparation(lobby LobbyPlan) float64 {
	team1, team2 := splitRosterByTeamNumber(lobby.Roster)
	return teamAverageSeparation(team1, team2)
}

// indexCombinations returns every k-length increasing index list over 0..n-1.
func indexCombinations(n, k int) [][]int {
	if k <= 0 || k > n {
		return nil
	}
	var out [][]int
	combo := make([]int, k)
	var rec func(start, filled int)
	rec = func(start, filled int) {
		if filled == k {
			cp := make([]int, k)
			copy(cp, combo)
			out = append(out, cp)
			return
		}
		for i := start; i <= n-(k-filled); i++ {
			combo[filled] = i
			rec(i+1, filled+1)
		}
	}
	rec(0, 0)
	return out
}
