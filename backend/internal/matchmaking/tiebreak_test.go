package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCompareByRankThenAvailability_HigherRankFirst(t *testing.T) {
	high := matchmaking.Player{UserID: uuid.New(), AvgRank: 20, CreatedAt: time.Now()}
	low := matchmaking.Player{UserID: uuid.New(), AvgRank: 10, CreatedAt: time.Now()}
	assert.Less(t, matchmaking.CompareByRankThenAvailability(high, low), 0)
	assert.Greater(t, matchmaking.CompareByRankThenAvailability(low, high), 0)
}

func TestCompareByRankThenAvailability_FewerGameRegistrationsWinOnTie(t *testing.T) {
	now := time.Now()
	fewerGames := matchmaking.Player{UserID: uuid.New(), AvgRank: 12, RegisteredGameCount: 1, CreatedAt: now}
	moreGames := matchmaking.Player{UserID: uuid.New(), AvgRank: 12, RegisteredGameCount: 2, CreatedAt: now}
	assert.Less(t, matchmaking.CompareByRankThenAvailability(fewerGames, moreGames), 0)
	assert.Greater(t, matchmaking.CompareByRankThenAvailability(moreGames, fewerGames), 0)
}

func TestCompareByRankThenAvailability_EarlierRegistrationWinsOnTie(t *testing.T) {
	now := time.Now()
	earlier := matchmaking.Player{UserID: uuid.New(), AvgRank: 12, RegisteredGameCount: 1, CreatedAt: now}
	later := matchmaking.Player{
		UserID:              uuid.New(),
		AvgRank:             12,
		RegisteredGameCount: 1,
		CreatedAt:           now.Add(time.Minute),
	}
	assert.Less(t, matchmaking.CompareByRankThenAvailability(earlier, later), 0)
	assert.Greater(t, matchmaking.CompareByRankThenAvailability(later, earlier), 0)
}

func TestCompareByRankThenAvailability_FullyEqual(t *testing.T) {
	now := time.Now()
	a := matchmaking.Player{UserID: uuid.New(), AvgRank: 12, RegisteredGameCount: 1, CreatedAt: now}
	b := matchmaking.Player{UserID: uuid.New(), AvgRank: 12, RegisteredGameCount: 1, CreatedAt: now}
	assert.Equal(t, 0, matchmaking.CompareByRankThenAvailability(a, b))
}

func TestCompareByRankThenAvailability_CanSubstituteDoesNotAffectOrder(t *testing.T) {
	now := time.Now()
	sub := matchmaking.Player{UserID: uuid.New(), AvgRank: 12, CanSubstitute: true, CreatedAt: now}
	nonSub := matchmaking.Player{UserID: uuid.New(), AvgRank: 12, CanSubstitute: false, CreatedAt: now}
	assert.Equal(t, 0, matchmaking.CompareByRankThenAvailability(sub, nonSub))
}
