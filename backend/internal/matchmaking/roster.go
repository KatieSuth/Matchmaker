package matchmaking

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
