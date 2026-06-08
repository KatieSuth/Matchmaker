package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssignBalanced_SnakeDraftAcrossLobbies(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 8)
	for i := 0; i < 8; i++ {
		players = append(players, matchmaking.Player{
			UserID:    uuid.New(),
			AvgRank:   float64(20 - i),
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		})
	}

	lobbies := matchmaking.AssignBalanced(players, 2, 4)
	require.Len(t, lobbies, 2)
	assert.Len(t, lobbies[0].Roster, 4)
	assert.Len(t, lobbies[1].Roster, 4)

	// Snake: lobby 0 gets ranks 20,17,16,13; lobby 1 gets 19,18,15,14.
	assert.Equal(t, 20.0, lobbies[0].Roster[0].AvgRank)
	assert.Equal(t, 19.0, lobbies[1].Roster[0].AvgRank)
}
