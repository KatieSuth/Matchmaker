package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func balancedCfg(teamSize, lobbyCount, tierCount int) matchmaking.Config {
	return matchmaking.Config{
		TeamSize:   teamSize,
		LobbyCount: lobbyCount,
		TierCount:  tierCount,
		Slots:      teamSize * 2,
		SortLogic:  "balanced",
	}
}

func TestAssignBalanced_IdealTwoWindowsOnePerTeam(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{
		playerAtRank(21, false, 1, now),
		playerAtRank(22, false, 1, now.Add(time.Minute)),
		playerAtRank(3, false, 1, now.Add(2*time.Minute)),
		playerAtRank(4, false, 1, now.Add(3*time.Minute)),
	}

	lobbies := matchmaking.AssignBalanced(players, balancedCfg(2, 1, 25))
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 4)

	var high, low int
	for _, p := range lobbies[0].Roster {
		if p.AvgRank >= 14 {
			high++
		} else {
			low++
		}
	}
	assert.Equal(t, 2, high)
	assert.Equal(t, 2, low)
}

func TestAssignBalanced_EmptyWindowTakesFourFromFullest(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 10)
	for i := 0; i < 10; i++ {
		players = append(players, playerAtRank(12, false, 1, now.Add(time.Duration(i)*time.Minute)))
	}

	lobbies := matchmaking.AssignBalanced(players, balancedCfg(5, 1, 25))
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 10)
	for _, p := range lobbies[0].Roster {
		assert.Equal(t, 12.0, p.AvgRank)
	}
}

func TestAssignBalanced_TwoEmptyWindowsTakeFourAndFour(t *testing.T) {
	now := time.Now()
	var players []matchmaking.Player
	for i := 0; i < 6; i++ {
		players = append(players, playerAtRank(3, false, 1, now.Add(time.Duration(i)*time.Minute)))
	}
	for i := 0; i < 4; i++ {
		players = append(players, playerAtRank(12, false, 1, now.Add(time.Duration(10+i)*time.Minute)))
	}

	lobbies := matchmaking.AssignBalanced(players, balancedCfg(5, 1, 25))
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 10)
	low, mid := 0, 0
	for _, p := range lobbies[0].Roster {
		if p.AvgRank == 3 {
			low++
		}
		if p.AvgRank == 12 {
			mid++
		}
	}
	assert.Equal(t, 6, low)
	assert.Equal(t, 4, mid)
}

func TestAssignBalanced_DonorWithThreeThenNextFullest(t *testing.T) {
	now := time.Now()
	var players []matchmaking.Player
	for i := 0; i < 7; i++ {
		players = append(players, playerAtRank(3, false, 1, now.Add(time.Duration(i)*time.Minute)))
	}
	for i := 0; i < 3; i++ {
		players = append(players, playerAtRank(12, false, 1, now.Add(time.Duration(20+i)*time.Minute)))
	}

	lobbies := matchmaking.AssignBalanced(players, balancedCfg(5, 1, 25))
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 10)
	low, mid := 0, 0
	for _, p := range lobbies[0].Roster {
		if p.AvgRank == 3 {
			low++
		}
		if p.AvgRank == 12 {
			mid++
		}
	}
	assert.Equal(t, 7, low)
	assert.Equal(t, 3, mid)
}

func TestAssignBalanced_OverflowSingletonBenched(t *testing.T) {
	now := time.Now()
	iron := playerAtRank(1, false, 1, now)
	players := []matchmaking.Player{iron}
	for i := 0; i < 14; i++ {
		players = append(players, playerAtRank(11, false, 1, now.Add(time.Duration(i+1)*time.Minute)))
	}

	lobbies := matchmaking.AssignBalanced(players, balancedCfg(5, 1, 25))
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 10)
	for _, p := range lobbies[0].Roster {
		assert.NotEqual(t, iron.UserID, p.UserID)
		assert.Equal(t, 11.0, p.AvgRank)
	}
}

func TestAssignBalanced_OverflowKeepsWindowMidpoint(t *testing.T) {
	now := time.Now()
	radiant := playerAtRank(25, false, 1, now)
	iron := playerAtRank(1, false, 1, now.Add(time.Minute))
	players := []matchmaking.Player{radiant, iron}
	for i := 0; i < 4; i++ {
		players = append(players, playerAtRank(21, false, 1, now.Add(time.Duration(i+2)*time.Minute)))
	}
	players = append(players, playerAtRank(8, false, 1, now.Add(10*time.Minute)))
	players = append(players, playerAtRank(6, false, 1, now.Add(11*time.Minute)))
	for i := 0; i < 8; i++ {
		players = append(players, playerAtRank(11, false, 1, now.Add(time.Duration(20+i)*time.Minute)))
	}

	lobbies := matchmaking.AssignBalanced(players, balancedCfg(2, 1, 25))
	require.Len(t, lobbies, 1)
	require.Len(t, lobbies[0].Roster, 4)
	ids := map[uuid.UUID]bool{}
	for _, p := range lobbies[0].Roster {
		ids[p.UserID] = true
	}
	assert.False(t, ids[radiant.UserID], "Radiant should sit when the high window has closer Ascendants")
	assert.False(t, ids[iron.UserID], "Iron should sit when the low window has closer Silver/Bronze")
}

func TestAssignBalanced_ParallelDealDoesNotFillFirstLobbyAlone(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 8)
	for i := 0; i < 4; i++ {
		players = append(players, playerAtRank(21, false, 1, now.Add(time.Duration(i)*time.Minute)))
	}
	for i := 0; i < 4; i++ {
		players = append(players, playerAtRank(3, false, 1, now.Add(time.Duration(10+i)*time.Minute)))
	}

	lobbies := matchmaking.AssignBalanced(players, balancedCfg(2, 2, 25))
	require.Len(t, lobbies, 2)
	assert.Len(t, lobbies[0].Roster, 4)
	assert.Len(t, lobbies[1].Roster, 4)
	highCounts := [2]int{}
	for i, lobby := range lobbies {
		for _, p := range lobby.Roster {
			if p.AvgRank >= 14 {
				highCounts[i]++
			}
		}
	}
	assert.Equal(t, 2, highCounts[0])
	assert.Equal(t, 2, highCounts[1])
}

func TestAssignBalanced_SkipsWindowWhenLobbyAlreadyFull(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{
		playerAtRank(21, false, 1, now),
		playerAtRank(22, false, 1, now.Add(time.Minute)),
		playerAtRank(3, false, 1, now.Add(2*time.Minute)),
		playerAtRank(4, false, 1, now.Add(3*time.Minute)),
	}
	lobbies := matchmaking.AssignBalanced(players, matchmaking.Config{
		TeamSize:   2,
		LobbyCount: 1,
		TierCount:  25,
		Slots:      2,
	})
	require.Len(t, lobbies, 1)
	assert.Len(t, lobbies[0].Roster, 2)
}

func TestAssignBalanced_FallsBackToSnakeWhenTierCountInvalid(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 4)
	for i := 0; i < 4; i++ {
		players = append(players, playerAtRank(float64(10+i), false, 1, now.Add(time.Duration(i)*time.Minute)))
	}

	lobbies := matchmaking.AssignBalanced(players, matchmaking.Config{
		TeamSize:   2,
		LobbyCount: 1,
		TierCount:  0,
		Slots:      4,
	})
	require.Len(t, lobbies, 1)
	assert.Len(t, lobbies[0].Roster, 4)
}

func TestAssignBalanced_ZeroLobbyCount(t *testing.T) {
	assert.Nil(t, matchmaking.AssignBalanced([]matchmaking.Player{{AvgRank: 10}}, matchmaking.Config{
		TeamSize:   2,
		LobbyCount: 0,
		TierCount:  25,
		Slots:      4,
	}))
}

func TestAssignBalanced_ZeroSlotsUsesTeamSize(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{
		playerAtRank(21, false, 1, now),
		playerAtRank(22, false, 1, now.Add(time.Minute)),
		playerAtRank(3, false, 1, now.Add(2*time.Minute)),
		playerAtRank(4, false, 1, now.Add(3*time.Minute)),
	}
	lobbies := matchmaking.AssignBalanced(players, matchmaking.Config{
		TeamSize:   2,
		LobbyCount: 1,
		TierCount:  25,
		Slots:      0,
	})
	require.Len(t, lobbies, 1)
	assert.Len(t, lobbies[0].Roster, 4)
}

func TestAssignBalanced_FallsBackToSnakeWhenTeamSizeInvalid(t *testing.T) {
	now := time.Now()
	low := playerAtRank(1, false, 1, now)
	high := playerAtRank(25, false, 1, now.Add(time.Minute))
	mid := playerAtRank(12, false, 1, now.Add(2*time.Minute))
	extra := playerAtRank(18, false, 1, now.Add(3*time.Minute))
	lobbies := matchmaking.AssignBalanced([]matchmaking.Player{low, high, mid, extra}, matchmaking.Config{
		TeamSize:   0,
		LobbyCount: 2,
		TierCount:  25,
		Slots:      2,
	})
	require.Len(t, lobbies, 2)
	assert.Len(t, lobbies[0].Roster, 2)
	assert.Len(t, lobbies[1].Roster, 2)
}

func TestAssignBalanced_OddSlotTakesSingleFromFullest(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{
		playerAtRank(21, false, 1, now),
		playerAtRank(22, false, 1, now.Add(time.Minute)),
		playerAtRank(3, false, 1, now.Add(2*time.Minute)),
	}
	lobbies := matchmaking.AssignBalanced(players, matchmaking.Config{
		TeamSize:   2,
		LobbyCount: 1,
		TierCount:  25,
		Slots:      3,
	})
	require.Len(t, lobbies, 1)
	assert.Len(t, lobbies[0].Roster, 3)
}

func TestAssignBalanced_ShortLobbiesWhenPoolExhausted(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{
		playerAtRank(21, false, 1, now),
		playerAtRank(3, false, 1, now.Add(time.Minute)),
	}
	lobbies := matchmaking.AssignBalanced(players, balancedCfg(2, 2, 25))
	require.Len(t, lobbies, 2)
	assert.GreaterOrEqual(t, len(lobbies[0].Roster)+len(lobbies[1].Roster), 2)
}

func TestAssignBalancedSnake_TrimsPoolAndReversesOddRound(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 6)
	for i := 0; i < 6; i++ {
		players = append(players, playerAtRank(float64(i+1), false, 1, now.Add(time.Duration(i)*time.Minute)))
	}
	lobbies := matchmaking.AssignBalancedSnakeForTest(players, 2, 2)
	require.Len(t, lobbies, 2)
	assert.Len(t, lobbies[0].Roster, 2)
	assert.Len(t, lobbies[1].Roster, 2)
}

func TestAssignBalancedSnake_EmptyWhenSlotsInvalid(t *testing.T) {
	lobbies := matchmaking.AssignBalancedSnakeForTest([]matchmaking.Player{{AvgRank: 10}}, 2, 0)
	require.Len(t, lobbies, 2)
	assert.Empty(t, lobbies[0].Roster)
	assert.Empty(t, lobbies[1].Roster)
}

func TestAssignBalancedSnake_ZeroLobbyCount(t *testing.T) {
	lobbies := matchmaking.AssignBalancedSnakeForTest([]matchmaking.Player{{AvgRank: 10}}, 0, 4)
	assert.Empty(t, lobbies)
}

func TestSnakeLobbyIndex(t *testing.T) {
	assert.Equal(t, 0, matchmaking.SnakeLobbyIndexForTest(0, 0))
	assert.Equal(t, 0, matchmaking.SnakeLobbyIndexForTest(0, 2))
	assert.Equal(t, 1, matchmaking.SnakeLobbyIndexForTest(1, 2))
	assert.Equal(t, 1, matchmaking.SnakeLobbyIndexForTest(2, 2), "odd round snakes back")
	assert.Equal(t, 0, matchmaking.SnakeLobbyIndexForTest(3, 2))
}
