package matchmaking

// CompareByRankThenAvailability orders players for roster selection.
// Higher avg rank first; on tie, fewer group game registrations wins; then earlier registration.
func CompareByRankThenAvailability(a, b Player) int {
	if a.AvgRank > b.AvgRank {
		return -1
	}
	if a.AvgRank < b.AvgRank {
		return 1
	}
	if a.RegisteredGameCount < b.RegisteredGameCount {
		return -1
	}
	if a.RegisteredGameCount > b.RegisteredGameCount {
		return 1
	}
	if a.CreatedAt.Before(b.CreatedAt) {
		return -1
	}
	if a.CreatedAt.After(b.CreatedAt) {
		return 1
	}
	return 0
}
