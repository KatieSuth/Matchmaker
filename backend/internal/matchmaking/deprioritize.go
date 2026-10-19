package matchmaking

import (
	"bytes"
	"sort"
)

// markDeprioritized returns copies of players with Deprioritized set for anyone whose
// 6-month no-show count at this host meets cfg.NoShowThreshold. Held-out players are then
// added back, fewest no-shows first, until the eligible pool can fill every roster seat
// for cfg.LobbyCount lobbies. Subs are drawn from everyone, so only roster seats count.
// A threshold of 0 (feature off) clears every flag and changes nothing.
func markDeprioritized(players []Player, cfg Config) []Player {
	out := make([]Player, len(players))
	copy(out, players)
	for i := range out {
		out[i].Deprioritized = false
	}
	if cfg.NoShowThreshold <= 0 {
		return out
	}

	var held []int
	subTotal := 0
	eligible, eligibleNonSubs := 0, 0
	for i := range out {
		if out[i].CanSubstitute {
			subTotal++
		}
		if out[i].NoShowCount >= cfg.NoShowThreshold {
			out[i].Deprioritized = true
			held = append(held, i)
			continue
		}
		eligible++
		if !out[i].CanSubstitute {
			eligibleNonSubs++
		}
	}
	if len(held) == 0 {
		return out
	}
	sort.SliceStable(held, func(a, b int) bool {
		return addBackBefore(out[held[a]], out[held[b]])
	})

	seats := cfg.LobbyCount * cfg.Slots
	// With n >= 2 only so many can-subs may sit on rosters (the rest are reserved for
	// sub_min), so the non-sub share of the seats must be covered by non-subs.
	neededNonSubs := seats - MaxSubstituteEligibleOnRoster(subTotal, cfg.LobbyCount, cfg.SubMin)
	for _, idx := range held {
		if eligibleNonSubs >= neededNonSubs {
			break
		}
		if out[idx].CanSubstitute {
			continue
		}
		out[idx].Deprioritized = false
		eligible++
		eligibleNonSubs++
	}
	for _, idx := range held {
		if eligible >= seats {
			break
		}
		if !out[idx].Deprioritized {
			continue
		}
		out[idx].Deprioritized = false
		eligible++
	}
	return out
}

// addBackBefore orders held-out players for add-back: fewest no-shows first, then the
// availability tie-break without rank (fewer group games, earlier registration), then user ID.
func addBackBefore(a, b Player) bool {
	if a.NoShowCount != b.NoShowCount {
		return a.NoShowCount < b.NoShowCount
	}
	if a.RegisteredGameCount != b.RegisteredGameCount {
		return a.RegisteredGameCount < b.RegisteredGameCount
	}
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.Before(b.CreatedAt)
	}
	return bytes.Compare(a.UserID[:], b.UserID[:]) < 0
}

// eligiblePlayers drops deprioritized players for the initial lobby fill.
func eligiblePlayers(players []Player) []Player {
	out := make([]Player, 0, len(players))
	for _, p := range players {
		if !p.Deprioritized {
			out = append(out, p)
		}
	}
	return out
}

// deprioritizationRank orders repair candidates on a full tie: anyone not deprioritized
// first, then deprioritized players with fewer no-shows.
func deprioritizationRank(p Player) int {
	if !p.Deprioritized {
		return 0
	}
	return 1 + p.NoShowCount
}
