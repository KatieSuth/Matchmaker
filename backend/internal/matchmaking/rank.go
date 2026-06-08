package matchmaking

// AverageRankOrder returns the numeric skill value used for sorting and fairness.
func AverageRankOrder(currentOrder, peakOrder int) float64 {
	return (float64(currentOrder) + float64(peakOrder)) / 2.0
}
