package matchmaking_test

import (
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/stretchr/testify/assert"
)

func TestAverageRankOrder(t *testing.T) {
	assert.Equal(t, 18.5, matchmaking.AverageRankOrder(17, 20))
	assert.Equal(t, 19.0, matchmaking.AverageRankOrder(18, 20))
}

func TestFlooredAverageRankOrder(t *testing.T) {
	assert.Equal(t, 18, matchmaking.FlooredAverageRankOrder(17, 20))
	assert.Equal(t, 19, matchmaking.FlooredAverageRankOrder(18, 20))
	assert.Equal(t, 10, matchmaking.FlooredAverageRankOrder(10, 10))
}
