package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanEvent_SingleLobby(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 4)
	for i := 0; i < 4; i++ {
		players = append(players, matchmaking.Player{
			UserID:    uuid.New(),
			AvgRank:   float64(i + 1),
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		})
	}

	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    0,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}

	plan, err := matchmaking.PlanEvent(players, cfg, settings)
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)
	assert.Len(t, plan.Lobbies[0].Roster, 4)
	assert.False(t, plan.SubCapacityAdjusted)
}

func TestPlanEvent_RankedMode(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 4)
	for i := 0; i < 4; i++ {
		players = append(players, matchmaking.Player{
			UserID:    uuid.New(),
			AvgRank:   float64(i + 1),
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		})
	}

	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    0,
		SortLogic: "ranked",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}

	plan, err := matchmaking.PlanEvent(players, cfg, settings)
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)
	assert.NotNil(t, plan.Lobbies[0].HostID)
}

func TestPlanEvent_MultiLobbyWithMandatorySubs(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 11)
	for i := 0; i < 8; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       float64(i + 1),
			CanSubstitute: false,
			CreatedAt:     now.Add(time.Duration(i) * time.Minute),
		})
	}
	for i := 0; i < 3; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       float64(20 + i),
			CanSubstitute: true,
			CreatedAt:     now.Add(time.Duration(8+i) * time.Minute),
		})
	}

	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    1,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}

	plan, err := matchmaking.PlanEvent(players, cfg, settings)
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 2)
	assert.Len(t, plan.Lobbies[0].Subs, 1)
	assert.Len(t, plan.Lobbies[1].Subs, 1)
}

func TestPlanEvent_KeepsExtraLobbyWhenReservedCanSubRostersAreFair(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 18)
	for i := 0; i < 6; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       21,
			CanSubstitute: false,
			CreatedAt:     now.Add(time.Duration(i) * time.Minute),
		})
	}
	for i := 0; i < 6; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       8,
			CanSubstitute: false,
			CreatedAt:     now.Add(time.Duration(6+i) * time.Minute),
		})
	}
	for i := 0; i < 6; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       11,
			CanSubstitute: true,
			CreatedAt:     now.Add(time.Duration(12+i) * time.Minute),
		})
	}

	plan, err := matchmaking.PlanEvent(players, matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    2,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}, matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	})
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 3)
	for i, lobby := range plan.Lobbies {
		assert.False(t, lobby.FairnessWarning, "lobby %d should stay even with 2 high + 2 low non-subs", i)
	}
}

func TestPlanEvent_DropsReservedCanSubLobbyWhenNonSubRostersStayUnfair(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 18)
	players = append(players, matchmaking.Player{
		UserID:        uuid.New(),
		AvgRank:       25,
		CanSubstitute: false,
		CreatedAt:     now,
	})
	for i := 0; i < 11; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       1,
			CanSubstitute: false,
			CreatedAt:     now.Add(time.Duration(i+1) * time.Minute),
		})
	}
	for i := 0; i < 6; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       11,
			CanSubstitute: true,
			CreatedAt:     now.Add(time.Duration(12+i) * time.Minute),
		})
	}

	plan, err := matchmaking.PlanEvent(players, matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    2,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}, matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	})
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 2, "three all-Iron/Radiant lobbies stay unfair; drop to two")
}

func TestShouldDropReservedSubLobbyForTest(t *testing.T) {
	assert.False(t, matchmaking.ShouldDropReservedSubLobbyForTest(6, matchmaking.Config{LobbyCount: 1, SubMin: 2}))
	assert.False(t, matchmaking.ShouldDropReservedSubLobbyForTest(6, matchmaking.Config{LobbyCount: 2, SubMin: 0}))
	assert.False(t, matchmaking.ShouldDropReservedSubLobbyForTest(6, matchmaking.Config{LobbyCount: 2, SubMin: 2}))
	assert.True(t, matchmaking.ShouldDropReservedSubLobbyForTest(6, matchmaking.Config{LobbyCount: 3, SubMin: 2}))
	assert.True(t, matchmaking.ShouldDropReservedSubLobbyForTest(6, matchmaking.Config{LobbyCount: 2, SubMin: 3}))
}

func TestPlanEvent_InsufficientPlayers(t *testing.T) {
	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  5,
		SubMin:    0,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (5v5)",
		Slots:     10,
	}
	players := []matchmaking.Player{{UserID: uuid.New(), AvgRank: 10, CreatedAt: time.Now()}}

	_, err := matchmaking.PlanEvent(players, cfg, matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	})
	require.Error(t, err)
	var valErr *matchmaking.ValidationError
	require.ErrorAs(t, err, &valErr)
}

func TestPlanEvent_PreservesMutualDuoOnSameTeamWhenBalanceAllows(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()

	players := []matchmaking.Player{
		{UserID: a, DiscordName: "A", AvgRank: 20, CreatedAt: now},
		{UserID: b, DiscordName: "B", DuoRequest: duoRequest("C"), AvgRank: 18, CreatedAt: now.Add(time.Minute)},
		{UserID: c, DiscordName: "C", DuoRequest: duoRequest("B"), AvgRank: 17, CreatedAt: now.Add(2 * time.Minute)},
		{UserID: d, DiscordName: "D", AvgRank: 10, CreatedAt: now.Add(3 * time.Minute)},
	}

	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    0,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}

	plan, err := matchmaking.PlanEvent(players, cfg, settings)
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)

	teamFor := func(id uuid.UUID) int {
		for _, p := range plan.Lobbies[0].Roster {
			if p.UserID != id || p.TeamNumber == nil {
				continue
			}
			return *p.TeamNumber
		}
		return 0
	}
	assert.Equal(t, teamFor(b), teamFor(c))
}

func TestPlanEvent_TwoVTwoDoesNotRosterRadiantWithIronWhenCloserRanksExist(t *testing.T) {
	now := time.Now()
	radiant := uuid.New()
	ascA := uuid.New()
	ascB := uuid.New()
	goldA := uuid.New()
	silver := uuid.New()
	bronze := uuid.New()
	iron := uuid.New()

	players := []matchmaking.Player{
		{UserID: radiant, AvgRank: 25, CreatedAt: now},
		{UserID: ascA, AvgRank: 21, CreatedAt: now.Add(time.Minute)},
		{UserID: ascB, AvgRank: 21, CreatedAt: now.Add(2 * time.Minute)},
		{UserID: goldA, AvgRank: 11, CreatedAt: now.Add(5 * time.Minute)},
		{UserID: silver, AvgRank: 8, CreatedAt: now.Add(8 * time.Minute)},
		{UserID: bronze, AvgRank: 6, CreatedAt: now.Add(8*time.Minute + time.Second)},
		{UserID: iron, AvgRank: 1, CreatedAt: now.Add(9 * time.Minute)},
	}

	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    0,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}
	plan, err := matchmaking.PlanEvent(players, cfg, matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	})
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)
	require.Len(t, plan.Lobbies[0].Roster, 4)

	rostered := map[uuid.UUID]bool{}
	high, low := 0, 0
	for _, p := range plan.Lobbies[0].Roster {
		rostered[p.UserID] = true
		if p.AvgRank >= 14 {
			high++
		} else {
			low++
		}
		require.NotNil(t, p.TeamNumber)
	}
	assert.Equal(t, 2, high)
	assert.Equal(t, 2, low)
	assert.False(t, rostered[radiant])
	assert.False(t, rostered[iron])
}

func TestPlanEvent_SubCapacitySwapKeepsTwoVTwoWindows(t *testing.T) {
	now := time.Now()
	radiant := uuid.New()
	iron := uuid.New()
	players := make([]matchmaking.Player, 0, 14)
	for i := 0; i < 4; i++ {
		players = append(players, matchmaking.Player{
			UserID:    uuid.New(),
			AvgRank:   21,
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		})
	}
	players = append(players,
		matchmaking.Player{UserID: radiant, AvgRank: 25, CreatedAt: now.Add(10 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 18, CreatedAt: now.Add(11 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 6, CreatedAt: now.Add(12 * time.Minute)},
		matchmaking.Player{UserID: iron, AvgRank: 3, CreatedAt: now.Add(13 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 8, CanSubstitute: true, CreatedAt: now.Add(14 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 11, CanSubstitute: true, CreatedAt: now.Add(15 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 11, CanSubstitute: true, CreatedAt: now.Add(16 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 11, CanSubstitute: true, CreatedAt: now.Add(17 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 13, CanSubstitute: true, CreatedAt: now.Add(18 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 8, CanSubstitute: true, CreatedAt: now.Add(19 * time.Minute)},
	)

	plan, err := matchmaking.PlanEvent(players, matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    2,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}, matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	})
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 2)

	rostered := map[uuid.UUID]bool{}
	for i, lobby := range plan.Lobbies {
		high, low := 0, 0
		for _, p := range lobby.Roster {
			rostered[p.UserID] = true
			if p.AvgRank >= 14 {
				high++
			} else {
				low++
			}
		}
		assert.Equal(t, 2, high, "lobby %d should keep two high-window players", i)
		assert.Equal(t, 2, low, "lobby %d should keep two low-window players", i)
	}
	assert.False(t, rostered[radiant], "Radiant should not fill a low-window sub-capacity hole")
	assert.True(t, rostered[iron], "same-window Iron should replace the swapped can-sub")
	assert.False(t, plan.SubCapacityAdjusted, "do not warn the host when every lobby ends fair")
	for i, lobby := range plan.Lobbies {
		assert.False(t, lobby.FairnessWarning, "lobby %d should be fair after same-window replacement", i)
	}
}

func TestPlanEvent_RepairsUnfairLowWindowPairWithUnplacedIrons(t *testing.T) {
	now := time.Now()
	ironA := uuid.New()
	ironB := uuid.New()
	plat := uuid.New()
	bronze := uuid.New()

	players := make([]matchmaking.Player, 0, 16)
	for i := 0; i < 4; i++ {
		players = append(players, matchmaking.Player{
			UserID:    uuid.New(),
			AvgRank:   21,
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		})
	}
	players = append(players,
		matchmaking.Player{UserID: uuid.New(), AvgRank: 25, CreatedAt: now.Add(10 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 18, CreatedAt: now.Add(11 * time.Minute)},
		matchmaking.Player{UserID: bronze, AvgRank: 6, CreatedAt: now.Add(12 * time.Minute)},
		matchmaking.Player{UserID: ironA, AvgRank: 3, CreatedAt: now.Add(13 * time.Minute)},
		matchmaking.Player{UserID: ironB, AvgRank: 3, CreatedAt: now.Add(14 * time.Minute)},
		matchmaking.Player{UserID: plat, AvgRank: 13, CreatedAt: now.Add(15 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 8, CanSubstitute: true, CreatedAt: now.Add(16 * time.Minute)},
	)
	for i := 0; i < 5; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       11,
			CanSubstitute: true,
			CreatedAt:     now.Add(time.Duration(17+i) * time.Minute),
		})
	}

	plan, err := matchmaking.PlanEvent(players, matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    2,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}, matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	})
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 2)

	rostered := map[uuid.UUID]bool{}
	for _, lobby := range plan.Lobbies {
		for _, p := range lobby.Roster {
			rostered[p.UserID] = true
		}
	}
	assert.True(t, rostered[ironA] || rostered[ironB], "an unplaced Iron should repair the Plat-vs-Bronze split")
	assert.False(t, rostered[plat], "Plat 1 should come off the strong side")
	assert.False(t, plan.SubCapacityAdjusted, "repair made the lobbies fair, so the substitute warning stays off")
	for i, lobby := range plan.Lobbies {
		assert.False(t, lobby.FairnessWarning, "lobby %d should be fair after repair", i)
	}
}

func TestPlanEvent_RankedDoesNotRepairUnfairLowWindowPair(t *testing.T) {
	now := time.Now()
	iron := uuid.New()
	bronze := uuid.New()
	plat := uuid.New()
	ascA := uuid.New()
	ascB := uuid.New()

	players := []matchmaking.Player{
		{UserID: iron, AvgRank: 3, CreatedAt: now},
		{UserID: bronze, AvgRank: 6, CreatedAt: now.Add(time.Minute)},
		{UserID: plat, AvgRank: 13, CreatedAt: now.Add(2 * time.Minute)},
		{UserID: ascA, AvgRank: 21, CreatedAt: now.Add(3 * time.Minute)},
		{UserID: ascB, AvgRank: 21, CreatedAt: now.Add(4 * time.Minute)},
	}

	plan, err := matchmaking.PlanEvent(players, matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    0,
		SortLogic: "ranked",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}, matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	})
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)

	rostered := map[uuid.UUID]bool{}
	for _, p := range plan.Lobbies[0].Roster {
		rostered[p.UserID] = true
	}
	assert.True(t, rostered[plat], "ranked packing keeps the high contiguous band")
	assert.True(t, rostered[bronze])
	assert.True(t, rostered[ascA])
	assert.True(t, rostered[ascB])
	assert.False(t, rostered[iron], "ranked must not run the balanced leftover-Iron repair")
}

func TestPlanEvent_SubCapacityWarningStaysWhenLobbyRemainsUnfair(t *testing.T) {
	now := time.Now()
	plat := uuid.New()
	bronze := uuid.New()

	players := make([]matchmaking.Player, 0, 12)
	for i := 0; i < 4; i++ {
		players = append(players, matchmaking.Player{
			UserID:    uuid.New(),
			AvgRank:   21,
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		})
	}
	players = append(players,
		matchmaking.Player{UserID: bronze, AvgRank: 6, CreatedAt: now.Add(10 * time.Minute)},
		matchmaking.Player{UserID: plat, AvgRank: 13, CreatedAt: now.Add(11 * time.Minute)},
		matchmaking.Player{UserID: uuid.New(), AvgRank: 8, CanSubstitute: true, CreatedAt: now.Add(12 * time.Minute)},
	)
	for i := 0; i < 5; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       11,
			CanSubstitute: true,
			CreatedAt:     now.Add(time.Duration(13+i) * time.Minute),
		})
	}

	plan, err := matchmaking.PlanEvent(players, matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    2,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}, matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	})
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 2)

	anyUnfair := false
	for _, lobby := range plan.Lobbies {
		if lobby.FairnessWarning {
			anyUnfair = true
			break
		}
	}
	require.True(t, anyUnfair, "Plat-vs-Bronze should remain after sub-capacity with no leftover Irons")
	assert.True(t, plan.SubCapacityAdjusted)
}

func TestAnyLobbyUnfairForTest(t *testing.T) {
	assert.False(t, matchmaking.AnyLobbyUnfairForTest(nil))
	assert.False(t, matchmaking.AnyLobbyUnfairForTest([]matchmaking.LobbyPlan{{}, {}}))
	assert.True(t, matchmaking.AnyLobbyUnfairForTest([]matchmaking.LobbyPlan{
		{},
		{FairnessWarning: true},
	}))
}
