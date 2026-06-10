package matchmaking

import "math"

// AverageRankOrder returns the numeric skill value used for sorting and fairness.
func AverageRankOrder(currentOrder, peakOrder int) float64 {
	return (float64(currentOrder) + float64(peakOrder)) / 2.0
}

// FlooredAverageRankOrder returns the floored rank order used to resolve a stored avg_rank.
func FlooredAverageRankOrder(currentOrder, peakOrder int) int {
	return int(math.Floor(AverageRankOrder(currentOrder, peakOrder)))
}
