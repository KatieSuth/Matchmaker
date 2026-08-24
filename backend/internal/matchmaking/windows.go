package matchmaking

import (
	"math"

	"github.com/google/uuid"
)

// rankWindow is one skill band on the game ladder. Window count matches team size
// (or tier count, whichever is smaller) so each team can take one player per band.
type rankWindow struct {
	minOrder int
	maxOrder int
	midpoint float64
}

// windowLeftover is an unpaired player after even window pairs are assigned.
type windowLeftover struct {
	player Player
	window int
	order  int
}

// buildRankWindows splits 1..tierCount into min(teamSize, tierCount) contiguous bands.
// Remainder ranks go to the lowest windows so low skill is never squeezed into a thinner band.
func buildRankWindows(tierCount, teamSize int) []rankWindow {
	if tierCount <= 0 || teamSize <= 0 {
		return nil
	}
	n := teamSize
	if tierCount < n {
		n = tierCount
	}

	base := tierCount / n
	extra := tierCount % n
	windows := make([]rankWindow, n)
	order := 1
	for i := 0; i < n; i++ {
		size := base
		if i < extra {
			size++
		}
		maxOrder := order + size - 1
		windows[i] = rankWindow{
			minOrder: order,
			maxOrder: maxOrder,
			midpoint: float64(order+maxOrder) / 2.0,
		}
		order = maxOrder + 1
	}
	return windows
}

// clampedRankOrder maps a player's AvgRank onto the ladder, inclusive of 1..tierCount.
func clampedRankOrder(avgRank float64, tierCount int) int {
	order := int(math.Round(avgRank))
	if order < 1 {
		order = 1
	}
	if tierCount > 0 && order > tierCount {
		order = tierCount
	}
	return order
}

// windowIndexForOrder maps a ladder order onto a window. Ranks off either end of the
// ladder clamp to the first or last band so a bad AvgRank cannot panic later buckets.
func windowIndexForOrder(order int, windows []rankWindow) int {
	if len(windows) == 0 {
		return 0
	}
	if order < windows[0].minOrder {
		return 0
	}
	last := len(windows) - 1
	if order > windows[last].maxOrder {
		return last
	}
	for i, w := range windows {
		if order >= w.minOrder && order <= w.maxOrder {
			return i
		}
	}
	return last
}

// bucketPlayersByWindow groups players by rank window. Empty windows stay in the slice
// so later fills can still address every band by index.
func bucketPlayersByWindow(players []Player, windows []rankWindow, tierCount int) [][]Player {
	buckets := make([][]Player, len(windows))
	for _, p := range players {
		idx := windowIndexForOrder(clampedRankOrder(p.AvgRank, tierCount), windows)
		buckets[idx] = append(buckets[idx], p)
	}
	return buckets
}

// pickClosestToMidpoint takes k players nearest the window's ladder midpoint so overflow
// keeps typical ranks of that band rather than its outliers.
func pickClosestToMidpoint(players []Player, k int, midpoint float64) (picked, rest []Player) {
	if k <= 0 || len(players) == 0 {
		return nil, append([]Player(nil), players...)
	}
	if k >= len(players) {
		return append([]Player(nil), players...), nil
	}

	ordered := append([]Player(nil), players...)
	sortPlayersByMidpointDistance(ordered, midpoint)
	picked = append([]Player(nil), ordered[:k]...)
	pickedIDs := make(map[uuid.UUID]struct{}, k)
	for _, p := range picked {
		pickedIDs[p.UserID] = struct{}{}
	}
	rest = make([]Player, 0, len(players)-k)
	for _, p := range players {
		if _, ok := pickedIDs[p.UserID]; ok {
			continue
		}
		rest = append(rest, p)
	}
	return picked, rest
}

// sortPlayersByMidpointDistance insertion-sorts in place, nearest the window midpoint first.
func sortPlayersByMidpointDistance(players []Player, midpoint float64) {
	for i := 1; i < len(players); i++ {
		for j := i; j > 0; j-- {
			if !closerToMidpoint(players[j], players[j-1], midpoint) {
				break
			}
			players[j], players[j-1] = players[j-1], players[j]
		}
	}
}

// closerToMidpoint reports whether a is a better window pick than b. Equal distance
// uses CompareByRankThenAvailability so ties stay deterministic.
func closerToMidpoint(a, b Player, midpoint float64) bool {
	da := math.Abs(a.AvgRank - midpoint)
	db := math.Abs(b.AvgRank - midpoint)
	if da < db {
		return true
	}
	if da > db {
		return false
	}
	return CompareByRankThenAvailability(a, b) < 0
}

// fullestWindowIndex returns the band with the most leftover players so extras come
// from the deepest pool. On a count tie the lower-index (lower-skill) window wins.
func fullestWindowIndex(remaining [][]Player) int {
	best := -1
	bestCount := 0
	for i, bucket := range remaining {
		n := len(bucket)
		if n == 0 {
			continue
		}
		if best < 0 || n > bestCount {
			best = i
			bestCount = n
		}
	}
	return best
}

// assignToWeakerTeam puts the higher-ranked of a pair on the team with the lower
// rank sum so each window pair pulls the running totals back toward even.
func assignToWeakerTeam(higher, lower Player, team1, team2 *[]Player, sum1, sum2 *float64) {
	if *sum1 <= *sum2 {
		higher.TeamNumber = intPtr(1)
		lower.TeamNumber = intPtr(2)
		*team1 = append(*team1, higher)
		*team2 = append(*team2, lower)
		*sum1 += higher.AvgRank
		*sum2 += lower.AvgRank
		return
	}
	higher.TeamNumber = intPtr(2)
	lower.TeamNumber = intPtr(1)
	*team2 = append(*team2, higher)
	*team1 = append(*team1, lower)
	*sum2 += higher.AvgRank
	*sum1 += lower.AvgRank
}

// appendToWeakerTeam places a leftover odd player on the team that currently has
// the lower rank sum. Used when a window has no pair partner.
func appendToWeakerTeam(p Player, team1, team2 *[]Player, sum1, sum2 *float64) {
	if *sum1 <= *sum2 {
		p.TeamNumber = intPtr(1)
		*team1 = append(*team1, p)
		*sum1 += p.AvgRank
		return
	}
	p.TeamNumber = intPtr(2)
	*team2 = append(*team2, p)
	*sum2 += p.AvgRank
}

// windowDraftIntoTeams puts one player from each window on each team when the lobby
// has a clean pair per band. Extra players in a band are paired with neighbors; odd
// leftovers pair with the closest leftover match, not with a far outlier.
func windowDraftIntoTeams(roster []Player, teamSize, tierCount int) ([]Player, []Player) {
	if len(roster) == 0 {
		return nil, nil
	}
	if teamSize <= 0 || tierCount <= 0 {
		return snakeDraftIntoTeams(roster, teamSize)
	}
	windows := buildRankWindows(tierCount, teamSize)
	if len(windows) == 0 {
		return snakeDraftIntoTeams(roster, teamSize)
	}

	buckets := bucketPlayersByWindow(roster, windows, tierCount)
	team1 := make([]Player, 0, teamSize)
	team2 := make([]Player, 0, teamSize)
	var sum1, sum2 float64
	leftovers := make([]windowLeftover, 0)

	for w, bucket := range buckets {
		sorted := append([]Player(nil), bucket...)
		sortPlayersByRankDesc(sorted)
		for i := 0; i+1 < len(sorted); i += 2 {
			higher, lower := sorted[i], sorted[i+1]
			assignToWeakerTeam(higher, lower, &team1, &team2, &sum1, &sum2)
		}
		if len(sorted)%2 == 1 {
			odd := sorted[len(sorted)-1]
			leftovers = append(leftovers, windowLeftover{
				player: odd,
				window: w,
				order:  clampedRankOrder(odd.AvgRank, tierCount),
			})
		}
	}

	for len(leftovers) >= 2 {
		i, j := bestLeftoverPair(leftovers)
		a, b := leftovers[i], leftovers[j]
		higher, lower := a.player, b.player
		if CompareByRankThenAvailability(b.player, a.player) < 0 {
			higher, lower = b.player, a.player
		}
		assignToWeakerTeam(higher, lower, &team1, &team2, &sum1, &sum2)
		leftovers = removeLeftoverPair(leftovers, i, j)
	}
	if len(leftovers) == 1 {
		appendToWeakerTeam(leftovers[0].player, &team1, &team2, &sum1, &sum2)
	}
	return team1, team2
}

// bestLeftoverPair chooses two leftovers closest in rank so an odd Radiant is not
// paired with an odd Iron when nearer leftovers exist. Equal distance prefers
// adjacent windows.
func bestLeftoverPair(leftovers []windowLeftover) (int, int) {
	bestI, bestJ := 0, 1
	bestDist := math.Abs(float64(leftovers[0].order - leftovers[1].order))
	bestAdj := absInt(leftovers[0].window-leftovers[1].window) == 1
	for i := 0; i < len(leftovers); i++ {
		for j := i + 1; j < len(leftovers); j++ {
			dist := math.Abs(float64(leftovers[i].order - leftovers[j].order))
			adj := absInt(leftovers[i].window-leftovers[j].window) == 1
			if dist < bestDist || (dist == bestDist && adj && !bestAdj) {
				bestI, bestJ = i, j
				bestDist = dist
				bestAdj = adj
			}
		}
	}
	return bestI, bestJ
}

// removeLeftoverPair drops indices i and j. Order of i and j does not matter.
func removeLeftoverPair(leftovers []windowLeftover, i, j int) []windowLeftover {
	if i > j {
		i, j = j, i
	}
	out := make([]windowLeftover, 0, len(leftovers)-2)
	for k, item := range leftovers {
		if k == i || k == j {
			continue
		}
		out = append(out, item)
	}
	return out
}

// absInt is |v| for leftover window-distance checks.
func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// splitIntoTeamsWindowed assigns teams via rank windows, then applies the same duo
// post-pass ranked mode uses — balance still wins.
func splitIntoTeamsWindowed(roster []Player, teamSize, tierCount int) ([]Player, []Player) {
	team1, team2 := windowDraftIntoTeams(roster, teamSize, tierCount)
	if len(team1) == 0 && len(team2) == 0 {
		return team1, team2
	}
	baselineSep := teamAverageSeparation(team1, team2)
	return applyDuoTeamGrouping(team1, team2, baselineSep)
}
