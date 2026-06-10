package matchmaking

import (
	"math"

	"github.com/google/uuid"
)

// balancedPickPhase tracks which skill tier to take next during balanced roster trimming.
type balancedPickPhase int

const (
	balancedPickLow balancedPickPhase = iota
	balancedPickMid
	balancedPickHigh
)

// selectBalancedRosterPool chooses roster players without rank-first cutoff.
// When trimming is required, players are taken in a low → mid → high cycle so the
// roster spans skill levels without excluding the middle of the pool.
// can_substitute does not affect selection.
func selectBalancedRosterPool(players []Player, needed int) []Player {
	if needed >= len(players) {
		return append([]Player(nil), players...)
	}

	remaining := append([]Player(nil), players...)
	pool := make([]Player, 0, needed)
	phase := balancedPickLow

	for len(pool) < needed && len(remaining) > 0 {
		sortPlayersByRankDesc(remaining)

		idx := balancedPickIndex(remaining, phase)
		picked := remaining[idx]

		pool = append(pool, picked)
		remaining = removePlayerByID(remaining, picked.UserID)
		phase = (phase + 1) % 3
	}

	return pool
}

// balancedPickIndex returns the roster index to take for the current low/mid/high cycle phase.
func balancedPickIndex(remaining []Player, phase balancedPickPhase) int {
	n := len(remaining)
	switch phase {
	case balancedPickLow:
		return n - 1
	case balancedPickMid:
		return balancedMidIndex(remaining)
	default:
		return 0
	}
}

// balancedMidIndex picks the player whose skill is closest to the remaining pool average.
func balancedMidIndex(remaining []Player) int {
	if len(remaining) == 0 {
		return 0
	}

	var sum float64
	for _, p := range remaining {
		sum += p.AvgRank
	}
	mean := sum / float64(len(remaining))

	bestIdx := 0
	bestDistance := math.Abs(remaining[0].AvgRank - mean)
	for i := 1; i < len(remaining); i++ {
		distance := math.Abs(remaining[i].AvgRank - mean)
		if distance < bestDistance {
			bestIdx = i
			bestDistance = distance
			continue
		}
		if distance == bestDistance && CompareByRankThenAvailability(remaining[i], remaining[bestIdx]) < 0 {
			bestIdx = i
		}
	}
	return bestIdx
}

// selectRankedRosterPool chooses a skill band for rank-grouping mode.
// When trimming is required, the majority skill side of the pool is kept.
func selectRankedRosterPool(players []Player, needed int) []Player {
	sorted := append([]Player(nil), players...)
	sortPlayersByRankAsc(sorted)
	if needed >= len(sorted) {
		return sorted
	}

	if rosterMajorityFavorsHigh(sorted) {
		return sorted[len(sorted)-needed:]
	}
	return sorted[:needed]
}

// removePlayerByID returns a copy of players with the given user removed.
func removePlayerByID(players []Player, id uuid.UUID) []Player {
	out := make([]Player, 0, len(players)-1)
	for _, p := range players {
		if p.UserID != id {
			out = append(out, p)
		}
	}
	return out
}

// rosterMajorityFavorsHigh reports whether more registrants are above the pool mean skill than below it.
// Used by ranked mode to decide whether overflow keeps the high or low skill band.
func rosterMajorityFavorsHigh(sorted []Player) bool {
	if len(sorted) == 0 {
		return false
	}

	var sum float64
	for _, p := range sorted {
		sum += p.AvgRank
	}
	mean := sum / float64(len(sorted))

	lowCount, highCount := 0, 0
	for _, p := range sorted {
		if p.AvgRank < mean {
			lowCount++
		} else if p.AvgRank > mean {
			highCount++
		}
	}
	if lowCount > highCount {
		return false
	}
	if highCount > lowCount {
		return true
	}
	return false
}
