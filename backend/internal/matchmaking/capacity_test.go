package matchmaking_test

import (
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequiredPlayers(t *testing.T) {
	assert.Equal(t, 0, matchmaking.RequiredPlayers(0, 10, 3))
	assert.Equal(t, 0, matchmaking.RequiredPlayers(-1, 10, 3))
	assert.Equal(t, 10, matchmaking.RequiredPlayers(1, 10, 3))
	assert.Equal(t, 26, matchmaking.RequiredPlayers(2, 10, 3))
}

func TestMaxLobbies(t *testing.T) {
	assert.Equal(t, 0, matchmaking.MaxLobbies(5, 0, 10, 3))
	assert.Equal(t, 1, matchmaking.MaxLobbies(20, 0, 10, 3))
	assert.Equal(t, 2, matchmaking.MaxLobbies(26, 6, 10, 3), "exact sub_min reservation still fits two lobbies")
	assert.Equal(t, 2, matchmaking.MaxLobbies(26, 7, 10, 3), "one spare can-sub may sit on a roster")
	assert.Equal(t, 1, matchmaking.MaxLobbies(26, 5, 10, 3))
	assert.Equal(t, 2, matchmaking.MaxLobbies(20, 0, 10, 0), "sub_min 0 still allows multiple lobbies with no can-subs")
}

func TestMaxLobbies_TwoVTwoAllowsThirdLobbyWhenCapacityFits(t *testing.T) {
	// 18 players, 6 can-sub, 2v2, sub_min=2: n=3 fits; PlanEvent may still drop if those rosters stay unfair.
	assert.Equal(t, 3, matchmaking.MaxLobbies(18, 6, 4, 2))
}

func TestValidateCapacity_InsufficientPlayers(t *testing.T) {
	_, err := matchmaking.ValidateCapacity(5, 0, 10, 3, "Game 1 (5v5)")
	require.Error(t, err)
	var valErr *matchmaking.ValidationError
	require.ErrorAs(t, err, &valErr)
	assert.Contains(t, valErr.Message, "needs at least 10 players")
}

func TestMaxSubstituteEligibleOnRoster(t *testing.T) {
	assert.Equal(t, 2, matchmaking.MaxSubstituteEligibleOnRoster(8, 2, 3))
	assert.Equal(t, 0, matchmaking.MaxSubstituteEligibleOnRoster(6, 2, 3))
	assert.Equal(t, 8, matchmaking.MaxSubstituteEligibleOnRoster(8, 1, 3))
}

func TestValidateCapacity_ZeroRegistrations(t *testing.T) {
	n, err := matchmaking.ValidateCapacity(0, 0, 10, 3, "Game 1")
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestMaxLobbies_CapsBySubstituteCount(t *testing.T) {
	// 26 roster-eligible players could support 2 lobbies at 5v5, but only 5 subs when 6 are required.
	assert.Equal(t, 1, matchmaking.MaxLobbies(26, 5, 10, 3))
}
