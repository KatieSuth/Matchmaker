package matchmaking_test

import (
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/stretchr/testify/assert"
)

func TestAverageRankOrder(t *testing.T) {
	assert.Equal(t, 12.5, matchmaking.AverageRankOrder(10, 15))
	assert.Equal(t, 25.0, matchmaking.AverageRankOrder(25, 25))
}
