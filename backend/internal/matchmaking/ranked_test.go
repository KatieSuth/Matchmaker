package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssignRanked_PacksContiguousBands(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{
		{UserID: uuid.New(), AvgRank: 20, CreatedAt: now},
		{UserID: uuid.New(), AvgRank: 19, CreatedAt: now.Add(time.Minute)},
		{UserID: uuid.New(), AvgRank: 18, CreatedAt: now.Add(2 * time.Minute)},
		{UserID: uuid.New(), AvgRank: 17, CreatedAt: now.Add(3 * time.Minute)},
		{UserID: uuid.New(), AvgRank: 5, CreatedAt: now.Add(4 * time.Minute)},
		{UserID: uuid.New(), AvgRank: 4, CreatedAt: now.Add(5 * time.Minute)},
		{UserID: uuid.New(), AvgRank: 3, CreatedAt: now.Add(6 * time.Minute)},
		{UserID: uuid.New(), AvgRank: 2, CreatedAt: now.Add(7 * time.Minute)},
	}

	lobbies := matchmaking.AssignRanked(players, 2, 4)
	require.Len(t, lobbies, 2)
	require.Len(t, lobbies[0].Roster, 4)
	require.Len(t, lobbies[1].Roster, 4)

	// Ascending pool: lowest ranks fill lobby 1 first.
	assert.Equal(t, 2.0, lobbies[0].Roster[0].AvgRank)
	assert.Equal(t, 5.0, lobbies[0].Roster[3].AvgRank)
	assert.Equal(t, 17.0, lobbies[1].Roster[0].AvgRank)
	assert.Equal(t, 20.0, lobbies[1].Roster[3].AvgRank)
}
