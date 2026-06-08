package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitIntoTeams_EmptyRoster(t *testing.T) {
	team1, team2 := matchmaking.SplitIntoTeams(nil, 2)
	assert.Nil(t, team1)
	assert.Nil(t, team2)
}

func TestSplitIntoTeams_SnakeDraftAssignsTeamNumbers(t *testing.T) {
	now := time.Now()
	roster := []matchmaking.Player{
		{UserID: uuid.New(), AvgRank: 20, CreatedAt: now},
		{UserID: uuid.New(), AvgRank: 18, CreatedAt: now.Add(time.Minute)},
		{UserID: uuid.New(), AvgRank: 16, CreatedAt: now.Add(2 * time.Minute)},
		{UserID: uuid.New(), AvgRank: 14, CreatedAt: now.Add(3 * time.Minute)},
	}

	team1, team2 := matchmaking.SplitIntoTeams(roster, 2)
	require.Len(t, team1, 2)
	require.Len(t, team2, 2)
	assert.Equal(t, 1, *team1[0].TeamNumber)
	assert.Equal(t, 1, *team1[1].TeamNumber)
	assert.Equal(t, 2, *team2[0].TeamNumber)
	assert.Equal(t, 2, *team2[1].TeamNumber)
}
