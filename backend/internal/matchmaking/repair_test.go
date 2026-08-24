package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func repairSettings() matchmaking.Settings {
	return matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}
}

func windowedLobby(t *testing.T, roster []matchmaking.Player, teamSize, tierCount int) matchmaking.LobbyPlan {
	t.Helper()
	team1, team2 := matchmaking.SplitIntoTeamsWindowedForTest(roster, teamSize, tierCount)
	return matchmaking.LobbyPlan{Roster: append(append([]matchmaking.Player{}, team1...), team2...)}
}

func rosterIDSet(lobby matchmaking.LobbyPlan) map[uuid.UUID]bool {
	ids := make(map[uuid.UUID]bool, len(lobby.Roster))
	for _, p := range lobby.Roster {
		ids[p.UserID] = true
	}
	return ids
}

func TestRepairUnfairWindowPairs_ReplacesPlatBronzeWithUnplacedIrons(t *testing.T) {
	now := time.Now()
	ascA := player(uuid.New(), 21, false, now)
	ascB := player(uuid.New(), 21, false, now.Add(time.Minute))
	plat := player(uuid.New(), 13, false, now.Add(2*time.Minute))
	bronze := player(uuid.New(), 6, false, now.Add(3*time.Minute))
	ironA := player(uuid.New(), 3, false, now.Add(4*time.Minute))
	ironB := player(uuid.New(), 3, false, now.Add(5*time.Minute))

	lobby := windowedLobby(t, []matchmaking.Player{ascA, ascB, plat, bronze}, 2, 25)
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		[]matchmaking.Player{ascA, ascB, plat, bronze, ironA, ironB},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	require.Len(t, out, 1)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[ascA.UserID])
	assert.True(t, ids[ascB.UserID])
	assert.True(t, ids[ironA.UserID] || ids[ironB.UserID], "one leftover Iron is enough in-place")
	assert.False(t, ids[plat.UserID])
	assert.True(t, ids[bronze.UserID], "in-place keeps the weak-side Bronze")

	team1, team2 := matchmaking.SplitRosterByTeamNumberForTest(out[0].Roster)
	assert.Less(t, matchmaking.TeamAverageSeparationForTest(team1, team2), 3.0)
}

func TestRepairUnfairWindowPairs_DoesNotTakeReservedCanSubs(t *testing.T) {
	now := time.Now()
	ascA := player(uuid.New(), 21, false, now)
	ascB := player(uuid.New(), 21, false, now.Add(time.Minute))
	plat := player(uuid.New(), 13, false, now.Add(2*time.Minute))
	bronze := player(uuid.New(), 6, false, now.Add(3*time.Minute))
	ironCS1 := player(uuid.New(), 3, true, now.Add(4*time.Minute))
	ironCS2 := player(uuid.New(), 3, true, now.Add(5*time.Minute))
	goldA := player(uuid.New(), 11, false, now.Add(6*time.Minute))
	goldB := player(uuid.New(), 11, false, now.Add(7*time.Minute))
	goldC := player(uuid.New(), 11, false, now.Add(8*time.Minute))
	goldD := player(uuid.New(), 11, false, now.Add(9*time.Minute))

	unfair := windowedLobby(t, []matchmaking.Player{ascA, ascB, plat, bronze}, 2, 25)
	fair := windowedLobby(t, []matchmaking.Player{goldA, goldB, goldC, goldD}, 2, 25)

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{unfair, fair},
		[]matchmaking.Player{ascA, ascB, plat, bronze, ironCS1, ironCS2, goldA, goldB, goldC, goldD},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 2, SubMin: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	require.Len(t, out, 2)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[plat.UserID])
	assert.True(t, ids[bronze.UserID])
	assert.False(t, ids[ironCS1.UserID])
	assert.False(t, ids[ironCS2.UserID])
}

func TestRepairUnfairWindowPairs_CanTakeOverflowCanSub(t *testing.T) {
	now := time.Now()
	ascA := player(uuid.New(), 21, false, now)
	ascB := player(uuid.New(), 21, false, now.Add(time.Minute))
	plat := player(uuid.New(), 13, false, now.Add(2*time.Minute))
	bronze := player(uuid.New(), 6, false, now.Add(3*time.Minute))
	ironCS1 := player(uuid.New(), 3, true, now.Add(4*time.Minute))
	ironCS2 := player(uuid.New(), 3, true, now.Add(5*time.Minute))
	ironCS3 := player(uuid.New(), 3, true, now.Add(6*time.Minute))
	goldA := player(uuid.New(), 11, false, now.Add(7*time.Minute))
	goldB := player(uuid.New(), 11, false, now.Add(8*time.Minute))
	goldC := player(uuid.New(), 11, false, now.Add(9*time.Minute))
	goldD := player(uuid.New(), 11, false, now.Add(10*time.Minute))

	unfair := windowedLobby(t, []matchmaking.Player{ascA, ascB, plat, bronze}, 2, 25)
	fair := windowedLobby(t, []matchmaking.Player{goldA, goldB, goldC, goldD}, 2, 25)

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{unfair, fair},
		[]matchmaking.Player{ascA, ascB, plat, bronze, ironCS1, ironCS2, ironCS3, goldA, goldB, goldC, goldD},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 2, SubMin: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	require.Len(t, out, 2)
	ids := rosterIDSet(out[0])
	taken := 0
	for _, id := range []uuid.UUID{ironCS1.UserID, ironCS2.UserID, ironCS3.UserID} {
		if ids[id] {
			taken++
		}
	}
	assert.Equal(t, 1, taken, "one spare can-sub may join; two would starve mandatory subs")
	assert.True(t, ids[bronze.UserID], "keeping Bronze plus one Iron is the legal improving pair")
	assert.False(t, ids[plat.UserID])
}

func TestRepairUnfairWindowPairs_DoesNotSwapAcrossWindows(t *testing.T) {
	now := time.Now()
	ascA := player(uuid.New(), 21, false, now)
	ascB := player(uuid.New(), 21, false, now.Add(time.Minute))
	plat := player(uuid.New(), 13, false, now.Add(2*time.Minute))
	bronze := player(uuid.New(), 6, false, now.Add(3*time.Minute))
	radiantA := player(uuid.New(), 25, false, now.Add(4*time.Minute))
	radiantB := player(uuid.New(), 25, false, now.Add(5*time.Minute))

	lobby := windowedLobby(t, []matchmaking.Player{ascA, ascB, plat, bronze}, 2, 25)
	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		[]matchmaking.Player{ascA, ascB, plat, bronze, radiantA, radiantB},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[plat.UserID], "high-window leftovers must not take a low-window slot")
	assert.True(t, ids[bronze.UserID], "high-window leftovers must not take a low-window slot")
	high, low := 0, 0
	for _, p := range out[0].Roster {
		if p.AvgRank >= 14 {
			high++
		} else {
			low++
		}
	}
	assert.Equal(t, 2, high)
	assert.Equal(t, 2, low)
}

func TestRepairUnfairWindowPairs_DoesNotStealFromOtherLobbies(t *testing.T) {
	now := time.Now()
	ascA := player(uuid.New(), 21, false, now)
	ascB := player(uuid.New(), 21, false, now.Add(time.Minute))
	plat := player(uuid.New(), 13, false, now.Add(2*time.Minute))
	bronze := player(uuid.New(), 6, false, now.Add(3*time.Minute))
	ironA := player(uuid.New(), 3, false, now.Add(4*time.Minute))
	ironB := player(uuid.New(), 3, false, now.Add(5*time.Minute))
	goldA := player(uuid.New(), 11, false, now.Add(6*time.Minute))
	goldB := player(uuid.New(), 11, false, now.Add(7*time.Minute))

	unfair := windowedLobby(t, []matchmaking.Player{ascA, ascB, plat, bronze}, 2, 25)
	other := windowedLobby(t, []matchmaking.Player{goldA, goldB, ironA, ironB}, 2, 25)

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{unfair, other},
		[]matchmaking.Player{ascA, ascB, plat, bronze, ironA, ironB, goldA, goldB},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 2, SortLogic: "balanced"},
		repairSettings(),
	)
	unfairIDs := rosterIDSet(out[0])
	otherIDs := rosterIDSet(out[1])
	assert.True(t, unfairIDs[plat.UserID])
	assert.True(t, unfairIDs[bronze.UserID])
	assert.True(t, otherIDs[ironA.UserID])
	assert.True(t, otherIDs[ironB.UserID])
}

func TestRepairUnfairWindowPairs_FairLobbyNoOp(t *testing.T) {
	now := time.Now()
	ascA := player(uuid.New(), 21, false, now)
	ascB := player(uuid.New(), 21, false, now.Add(time.Minute))
	goldA := player(uuid.New(), 11, false, now.Add(2*time.Minute))
	goldB := player(uuid.New(), 11, false, now.Add(3*time.Minute))
	ironA := player(uuid.New(), 3, false, now.Add(4*time.Minute))
	ironB := player(uuid.New(), 3, false, now.Add(5*time.Minute))

	lobby := windowedLobby(t, []matchmaking.Player{ascA, ascB, goldA, goldB}, 2, 25)
	require.False(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		[]matchmaking.Player{ascA, ascB, goldA, goldB, ironA, ironB},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[goldA.UserID])
	assert.True(t, ids[goldB.UserID])
	assert.False(t, ids[ironA.UserID])
	assert.False(t, ids[ironB.UserID])
}

func TestRepairUnfairWindowPairs_EmptyAndInvalidWindows(t *testing.T) {
	assert.Nil(t, matchmaking.RepairUnfairWindowPairsForTest(nil, nil, matchmaking.Config{}, repairSettings()))

	now := time.Now()
	lobby := windowedLobby(t, []matchmaking.Player{
		player(uuid.New(), 21, false, now),
		player(uuid.New(), 21, false, now.Add(time.Minute)),
		player(uuid.New(), 13, false, now.Add(2*time.Minute)),
		player(uuid.New(), 6, false, now.Add(3*time.Minute)),
	}, 2, 25)
	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		nil,
		matchmaking.Config{TeamSize: 0, TierCount: 25, SortLogic: "balanced"},
		repairSettings(),
	)
	require.Len(t, out, 1)
	assert.Len(t, out[0].Roster, 4)
}

func TestRepairUnfairWindowPairs_RankedSortLogicNoOp(t *testing.T) {
	now := time.Now()
	ascA := player(uuid.New(), 21, false, now)
	ascB := player(uuid.New(), 21, false, now.Add(time.Minute))
	plat := player(uuid.New(), 13, false, now.Add(2*time.Minute))
	bronze := player(uuid.New(), 6, false, now.Add(3*time.Minute))
	ironA := player(uuid.New(), 3, false, now.Add(4*time.Minute))
	ironB := player(uuid.New(), 3, false, now.Add(5*time.Minute))

	lobby := windowedLobby(t, []matchmaking.Player{ascA, ascB, plat, bronze}, 2, 25)
	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		[]matchmaking.Player{ascA, ascB, plat, bronze, ironA, ironB},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "ranked"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[plat.UserID])
	assert.True(t, ids[bronze.UserID])
	assert.False(t, ids[ironA.UserID])
}

func TestRepairUnfairWindowPairs_FiveVFiveReplacesOnlyBadWindow(t *testing.T) {
	now := time.Now()
	ironLow := player(uuid.New(), 1, false, now)
	bronzeHigh := player(uuid.New(), 5, false, now.Add(time.Minute))
	ironMidA := player(uuid.New(), 3, false, now.Add(2*time.Minute))
	ironMidB := player(uuid.New(), 3, false, now.Add(3*time.Minute))
	diamondLeftover := player(uuid.New(), 18, false, now.Add(4*time.Minute))

	w1a := player(uuid.New(), 10, false, now.Add(5*time.Minute))
	w1b := player(uuid.New(), 10, false, now.Add(6*time.Minute))
	w2a := player(uuid.New(), 15, false, now.Add(7*time.Minute))
	w2b := player(uuid.New(), 15, false, now.Add(8*time.Minute))
	w3a := player(uuid.New(), 20, false, now.Add(9*time.Minute))
	w3b := player(uuid.New(), 20, false, now.Add(10*time.Minute))
	w4a := player(uuid.New(), 25, false, now.Add(11*time.Minute))
	w4b := player(uuid.New(), 25, false, now.Add(12*time.Minute))

	roster := []matchmaking.Player{ironLow, bronzeHigh, w1a, w1b, w2a, w2b, w3a, w3b, w4a, w4b}
	lobby := windowedLobby(t, roster, 5, 25)
	settings := matchmaking.Settings{
		FairnessOutlierGap:         100,
		FairnessTeamSeparation:     0.5,
		FairnessReferenceTierCount: 25,
	}
	require.True(t, matchmaking.IsLobbyUnfair(lobby, settings, 25))

	all := append(append([]matchmaking.Player{}, roster...), ironMidA, ironMidB, diamondLeftover)
	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		all,
		matchmaking.Config{TeamSize: 5, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		settings,
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[ironMidA.UserID] || ids[ironMidB.UserID], "in-place pulls a same-window leftover")
	assert.False(t, ids[diamondLeftover.UserID])
	assert.True(t, ids[w3a.UserID])
	assert.True(t, ids[w3b.UserID])
	team1, team2 := matchmaking.SplitRosterByTeamNumberForTest(out[0].Roster)
	assert.Less(t, matchmaking.TeamAverageSeparationForTest(team1, team2), 0.5)
}

func TestRepairUnfairWindowPairs_RejectsWorseMixWhenBetterPairExists(t *testing.T) {
	now := time.Now()
	ascA := player(uuid.New(), 21, false, now)
	ascB := player(uuid.New(), 21, false, now.Add(time.Minute))
	plat := player(uuid.New(), 13, false, now.Add(2*time.Minute))
	bronze := player(uuid.New(), 6, false, now.Add(3*time.Minute))
	iron := player(uuid.New(), 3, false, now.Add(4*time.Minute))

	lobby := windowedLobby(t, []matchmaking.Player{ascA, ascB, plat, bronze}, 2, 25)
	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		[]matchmaking.Player{ascA, ascB, plat, bronze, iron},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[iron.UserID])
	assert.True(t, ids[bronze.UserID], "Bronze+Iron improves sep; Plat+Iron would worsen it")
	assert.False(t, ids[plat.UserID])
}

func TestRepairUnfairWindowPairs_CapsLargeLeftoverPool(t *testing.T) {
	now := time.Now()
	ascA := player(uuid.New(), 21, false, now)
	ascB := player(uuid.New(), 21, false, now.Add(time.Minute))
	plat := player(uuid.New(), 13, false, now.Add(2*time.Minute))
	bronze := player(uuid.New(), 6, false, now.Add(3*time.Minute))

	all := []matchmaking.Player{ascA, ascB, plat, bronze}
	var ironIDs []uuid.UUID
	for i := 0; i < 15; i++ {
		p := player(uuid.New(), 3, false, now.Add(time.Duration(4+i)*time.Minute))
		all = append(all, p)
		ironIDs = append(ironIDs, p.UserID)
	}

	lobby := windowedLobby(t, []matchmaking.Player{ascA, ascB, plat, bronze}, 2, 25)
	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		all,
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	onRoster := 0
	for _, id := range ironIDs {
		if ids[id] {
			onRoster++
		}
	}
	assert.Equal(t, 1, onRoster, "in-place takes one leftover; the weak-side Bronze stays")
	assert.False(t, ids[plat.UserID])
	assert.True(t, ids[bronze.UserID])
}

func TestIndexCombinations(t *testing.T) {
	combos := matchmaking.IndexCombinationsForTest(4, 2)
	require.Len(t, combos, 6)
	assert.Equal(t, []int{0, 1}, combos[0])
	assert.Equal(t, []int{2, 3}, combos[5])

	assert.Equal(t, [][]int{{0, 1, 2}}, matchmaking.IndexCombinationsForTest(3, 3))
	assert.Nil(t, matchmaking.IndexCombinationsForTest(5, 0))
	assert.Nil(t, matchmaking.IndexCombinationsForTest(2, 3))
	assert.Nil(t, matchmaking.IndexCombinationsForTest(0, 1))
}

func TestSamePlayerSetForTest(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	assert.False(t, matchmaking.SamePlayerSetForTest([]matchmaking.Player{{UserID: a}}, []uuid.UUID{a, b}))
	assert.False(t, matchmaking.SamePlayerSetForTest([]matchmaking.Player{{UserID: a}, {UserID: b}}, []uuid.UUID{a, uuid.New()}))
	assert.True(t, matchmaking.SamePlayerSetForTest([]matchmaking.Player{{UserID: a}, {UserID: b}}, []uuid.UUID{a, b}))
}

func withTeam(p matchmaking.Player, team int) matchmaking.Player {
	n := team
	p.TeamNumber = &n
	return p
}

func TestRepairUnfairWindowPairs_InPlaceSameWindowKeepsOtherTeams(t *testing.T) {
	now := time.Now()
	plat := withTeam(player(uuid.New(), 13, false, now), 1)
	asc1 := withTeam(player(uuid.New(), 21, false, now.Add(time.Minute)), 1)
	bronze := withTeam(player(uuid.New(), 6, false, now.Add(2*time.Minute)), 2)
	asc2 := withTeam(player(uuid.New(), 21, false, now.Add(3*time.Minute)), 2)
	gold := player(uuid.New(), 11, false, now.Add(4*time.Minute))

	lobby := matchmaking.LobbyPlan{Roster: []matchmaking.Player{plat, asc1, bronze, asc2}}
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		[]matchmaking.Player{plat, asc1, bronze, asc2, gold},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	require.Len(t, out, 1)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[gold.UserID])
	assert.True(t, ids[plat.UserID], "raising the weak-side Bronze with Gold 2 improves sep more than lowering Plat")
	assert.False(t, ids[bronze.UserID])
	assert.True(t, ids[asc1.UserID])
	assert.True(t, ids[asc2.UserID])

	byID := map[uuid.UUID]matchmaking.Player{}
	for _, p := range out[0].Roster {
		byID[p.UserID] = p
	}
	require.NotNil(t, byID[gold.UserID].TeamNumber)
	assert.Equal(t, 2, *byID[gold.UserID].TeamNumber)
	assert.Equal(t, 1, *byID[plat.UserID].TeamNumber)
	assert.Equal(t, 1, *byID[asc1.UserID].TeamNumber)
	assert.Equal(t, 2, *byID[asc2.UserID].TeamNumber)

	team1, team2 := matchmaking.SplitRosterByTeamNumberForTest(out[0].Roster)
	assert.Less(t, matchmaking.TeamAverageSeparationForTest(team1, team2), 3.0)
}

func TestRepairUnfairWindowPairs_InPlaceLowersStrongPlatWithGold2(t *testing.T) {
	now := time.Now()
	plat := withTeam(player(uuid.New(), 13, false, now), 1)
	t1 := []matchmaking.Player{
		withTeam(player(uuid.New(), 25, false, now.Add(time.Minute)), 1),
		withTeam(player(uuid.New(), 20, false, now.Add(2*time.Minute)), 1),
		plat,
		withTeam(player(uuid.New(), 9, false, now.Add(3*time.Minute)), 1),
		withTeam(player(uuid.New(), 5, false, now.Add(4*time.Minute)), 1),
	}
	t2 := []matchmaking.Player{
		withTeam(player(uuid.New(), 24, false, now.Add(5*time.Minute)), 2),
		withTeam(player(uuid.New(), 16, false, now.Add(6*time.Minute)), 2),
		withTeam(player(uuid.New(), 8, false, now.Add(7*time.Minute)), 2),
		withTeam(player(uuid.New(), 6, false, now.Add(8*time.Minute)), 2),
		withTeam(player(uuid.New(), 2, false, now.Add(9*time.Minute)), 2),
	}
	gold2 := player(uuid.New(), 11, false, now.Add(10*time.Minute))
	roster := append(append([]matchmaking.Player{}, t1...), t2...)
	lobby := matchmaking.LobbyPlan{Roster: roster}
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		append(append([]matchmaking.Player{}, roster...), gold2),
		matchmaking.Config{TeamSize: 5, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[gold2.UserID])
	assert.False(t, ids[plat.UserID])

	byID := map[uuid.UUID]matchmaking.Player{}
	for _, p := range out[0].Roster {
		byID[p.UserID] = p
	}
	require.NotNil(t, byID[gold2.UserID].TeamNumber)
	assert.Equal(t, 1, *byID[gold2.UserID].TeamNumber)
	for _, p := range roster {
		if p.UserID == plat.UserID {
			continue
		}
		require.NotNil(t, byID[p.UserID].TeamNumber)
		assert.Equal(t, *p.TeamNumber, *byID[p.UserID].TeamNumber)
	}
}

func TestRepairUnfairWindowPairs_AdjacentFringeGold1ForPlat1(t *testing.T) {
	now := time.Now()
	t1 := []matchmaking.Player{
		withTeam(player(uuid.New(), 25, false, now), 1),
		withTeam(player(uuid.New(), 21, false, now.Add(time.Minute)), 1),
		withTeam(player(uuid.New(), 15, false, now.Add(2*time.Minute)), 1),
		withTeam(player(uuid.New(), 13, false, now.Add(3*time.Minute)), 1),
		withTeam(player(uuid.New(), 8, false, now.Add(4*time.Minute)), 1),
	}
	t2 := []matchmaking.Player{
		withTeam(player(uuid.New(), 24, false, now.Add(5*time.Minute)), 2),
		withTeam(player(uuid.New(), 16, false, now.Add(6*time.Minute)), 2),
		withTeam(player(uuid.New(), 14, false, now.Add(7*time.Minute)), 2),
		withTeam(player(uuid.New(), 11, false, now.Add(8*time.Minute)), 2),
		withTeam(player(uuid.New(), 1, false, now.Add(9*time.Minute)), 2),
	}
	plat := t1[3]
	gold1 := player(uuid.New(), 10, false, now.Add(10*time.Minute))
	roster := append(append([]matchmaking.Player{}, t1...), t2...)
	lobby := matchmaking.LobbyPlan{Roster: roster}
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		append(append([]matchmaking.Player{}, roster...), gold1),
		matchmaking.Config{TeamSize: 5, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	require.Len(t, out, 1)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[gold1.UserID])
	assert.False(t, ids[plat.UserID])

	byID := map[uuid.UUID]matchmaking.Player{}
	for _, p := range out[0].Roster {
		byID[p.UserID] = p
	}
	require.NotNil(t, byID[gold1.UserID].TeamNumber)
	assert.Equal(t, 1, *byID[gold1.UserID].TeamNumber)
	for _, p := range roster {
		if p.UserID == plat.UserID {
			continue
		}
		require.NotNil(t, byID[p.UserID].TeamNumber)
		assert.Equal(t, *p.TeamNumber, *byID[p.UserID].TeamNumber, "everyone else keeps their seat")
	}
}

func TestRepairUnfairWindowPairs_AdjacentFringeRejectsFarNeighbor(t *testing.T) {
	now := time.Now()
	t1 := []matchmaking.Player{
		withTeam(player(uuid.New(), 25, false, now), 1),
		withTeam(player(uuid.New(), 21, false, now.Add(time.Minute)), 1),
		withTeam(player(uuid.New(), 15, false, now.Add(2*time.Minute)), 1),
		withTeam(player(uuid.New(), 13, false, now.Add(3*time.Minute)), 1),
		withTeam(player(uuid.New(), 8, false, now.Add(4*time.Minute)), 1),
	}
	t2 := []matchmaking.Player{
		withTeam(player(uuid.New(), 23, false, now.Add(5*time.Minute)), 2),
		withTeam(player(uuid.New(), 16, false, now.Add(6*time.Minute)), 2),
		withTeam(player(uuid.New(), 12, false, now.Add(7*time.Minute)), 2),
		withTeam(player(uuid.New(), 7, false, now.Add(8*time.Minute)), 2),
		withTeam(player(uuid.New(), 6, false, now.Add(9*time.Minute)), 2),
	}
	iron := player(uuid.New(), 1, false, now.Add(10*time.Minute))
	roster := append(append([]matchmaking.Player{}, t1...), t2...)
	lobby := matchmaking.LobbyPlan{Roster: roster}
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		append(append([]matchmaking.Player{}, roster...), iron),
		matchmaking.Config{TeamSize: 5, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.False(t, ids[iron.UserID], "Iron 1 is not adjacent-fringe with Plat 1")
	assert.True(t, ids[t1[3].UserID])
}

func TestAdjacentFringeSwapOKForTest(t *testing.T) {
	assert.True(t, matchmaking.AdjacentFringeSwapOKForTest(13, 10, 5, 25), "5v5 Plat 1 ↔ Gold 1")
	assert.False(t, matchmaking.AdjacentFringeSwapOKForTest(13, 11, 5, 25), "Gold 2 is same 5v5 window as Plat 1")
	assert.False(t, matchmaking.AdjacentFringeSwapOKForTest(13, 1, 5, 25), "Iron 1 is two windows below Plat 1")
	assert.False(t, matchmaking.AdjacentFringeSwapOKForTest(12, 25, 5, 25), "Radiant is not fringe to Gold")
	assert.True(t, matchmaking.AdjacentFringeSwapOKForTest(13, 14, 2, 25), "2v2 Plat 1 ↔ Diamond 1 at 13|14")
	assert.False(t, matchmaking.AdjacentFringeSwapOKForTest(13, 25, 2, 25), "2v2 Radiant is not fringe to Plat")
	assert.False(t, matchmaking.AdjacentFringeSwapOKForTest(13, 10, 5, 0), "no windows")
}

func TestRepairSwapAllowedForTest(t *testing.T) {
	assert.True(t, matchmaking.RepairSwapAllowedForTest(13, 11, 5, 25, true), "same-window Gold 2")
	assert.False(t, matchmaking.RepairSwapAllowedForTest(13, 10, 5, 25, true), "Gold 1 is the next window")
	assert.True(t, matchmaking.RepairSwapAllowedForTest(13, 10, 5, 25, false), "fringe Gold 1")
	assert.False(t, matchmaking.RepairSwapAllowedForTest(13, 11, 5, 25, false), "same window is not fringe")
}

func TestLeftoverCanJoinRosterForTest(t *testing.T) {
	assert.True(t, matchmaking.LeftoverCanJoinRosterForTest(matchmaking.Player{}, 0, 2))
	assert.True(t, matchmaking.LeftoverCanJoinRosterForTest(matchmaking.Player{CanSubstitute: true}, 3, 2))
	assert.False(t, matchmaking.LeftoverCanJoinRosterForTest(matchmaking.Player{CanSubstitute: true}, 2, 2))
}

func TestReplaceRosterPlayerKeepingTeamForTest(t *testing.T) {
	outID := uuid.New()
	keepID := uuid.New()
	inID := uuid.New()
	t1, t2 := 1, 2
	roster := []matchmaking.Player{
		{UserID: outID, AvgRank: 13, TeamNumber: &t1},
		{UserID: keepID, AvgRank: 6, TeamNumber: &t2},
	}
	got := matchmaking.ReplaceRosterPlayerKeepingTeamForTest(roster, outID, matchmaking.Player{UserID: inID, AvgRank: 11})
	require.Len(t, got, 2)
	assert.Equal(t, inID, got[0].UserID)
	require.NotNil(t, got[0].TeamNumber)
	assert.Equal(t, 1, *got[0].TeamNumber)
	assert.Equal(t, keepID, got[1].UserID)
	assert.Equal(t, 2, *got[1].TeamNumber)

	unchanged := matchmaking.ReplaceRosterPlayerKeepingTeamForTest(roster, uuid.New(), matchmaking.Player{UserID: inID})
	assert.Equal(t, outID, unchanged[0].UserID)

	noTeam := []matchmaking.Player{{UserID: outID, AvgRank: 13}}
	kept := matchmaking.ReplaceRosterPlayerKeepingTeamForTest(noTeam, outID, matchmaking.Player{UserID: inID})
	assert.Equal(t, outID, kept[0].UserID)
}

func TestRankSpreadForTest(t *testing.T) {
	assert.Equal(t, 0.0, matchmaking.RankSpreadForTest(nil))
	assert.Equal(t, 0.0, matchmaking.RankSpreadForTest([]float64{5}))
	assert.Equal(t, 7.0, matchmaking.RankSpreadForTest([]float64{13, 6}))
	assert.Equal(t, 4.0, matchmaking.RankSpreadForTest([]float64{20, 16, 18}))
}

func TestWindowSpreadWorsensForTest(t *testing.T) {
	highA := uuid.New()
	highB := uuid.New()
	roster := []matchmaking.Player{
		{UserID: highA, AvgRank: 21},
		{UserID: highB, AvgRank: 21},
	}
	assert.True(t, matchmaking.WindowSpreadWorsensForTest(roster, highB, matchmaking.Player{AvgRank: 25}, 2, 25))
	assert.False(t, matchmaking.WindowSpreadWorsensForTest(
		[]matchmaking.Player{{UserID: highA, AvgRank: 13}, {UserID: highB, AvgRank: 6}},
		highA,
		matchmaking.Player{AvgRank: 11},
		2,
		25,
	))
	assert.False(t, matchmaking.WindowSpreadWorsensForTest(
		[]matchmaking.Player{{UserID: highA, AvgRank: 13}},
		highA,
		matchmaking.Player{AvgRank: 11},
		5,
		25,
	))
}

func TestCapWindowSearchPoolForTest(t *testing.T) {
	now := time.Now()
	rostered := []matchmaking.Player{
		player(uuid.New(), 21, false, now),
		player(uuid.New(), 21, false, now.Add(time.Minute)),
	}
	small := []matchmaking.Player{player(uuid.New(), 25, false, now.Add(2*time.Minute))}
	got := matchmaking.CapWindowSearchPoolForTest(rostered, small, 19.5)
	require.Len(t, got, 3)

	var leftovers []matchmaking.Player
	for i := 0; i < 15; i++ {
		leftovers = append(leftovers, player(uuid.New(), 25, false, now.Add(time.Duration(3+i)*time.Minute)))
	}
	capped := matchmaking.CapWindowSearchPoolForTest(rostered, leftovers, 19.5)
	require.Len(t, capped, 12)
	assert.Equal(t, rostered[0].UserID, capped[0].UserID)
	assert.Equal(t, rostered[1].UserID, capped[1].UserID)

	hugeRoster := make([]matchmaking.Player, 12)
	for i := range hugeRoster {
		hugeRoster[i] = player(uuid.New(), 12, false, now.Add(time.Duration(i)*time.Minute))
	}
	onlyRoster := matchmaking.CapWindowSearchPoolForTest(hugeRoster, leftovers, 13)
	require.Len(t, onlyRoster, 12)
	assert.Equal(t, hugeRoster[0].UserID, onlyRoster[0].UserID)
}

func TestRepairUnfairWindowPairs_OutlierSwapsUnplacedGold2(t *testing.T) {
	now := time.Now()
	plat := withTeam(player(uuid.New(), 13, false, now), 1)
	iron := withTeam(player(uuid.New(), 3, false, now.Add(time.Minute)), 1)
	bronzeA := withTeam(player(uuid.New(), 6, false, now.Add(2*time.Minute)), 2)
	bronzeB := withTeam(player(uuid.New(), 6, false, now.Add(3*time.Minute)), 2)
	gold := player(uuid.New(), 11, false, now.Add(4*time.Minute))

	lobby := matchmaking.LobbyPlan{Roster: []matchmaking.Player{plat, iron, bronzeA, bronzeB}}
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))
	team1, team2 := matchmaking.SplitRosterByTeamNumberForTest(lobby.Roster)
	assert.Less(t, matchmaking.TeamAverageSeparationForTest(team1, team2), 3.0)
	assert.Greater(t, matchmaking.OutlierGapForTest(lobby.Roster), 6.0)

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		[]matchmaking.Player{plat, iron, bronzeA, bronzeB, gold},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	require.Len(t, out, 1)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[gold.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], repairSettings(), 25))
}

func TestRepairUnfairWindowPairs_OutlierDoesNotTakeFarLeftover(t *testing.T) {
	now := time.Now()
	plat := withTeam(player(uuid.New(), 13, false, now), 1)
	iron := withTeam(player(uuid.New(), 3, false, now.Add(time.Minute)), 1)
	bronzeA := withTeam(player(uuid.New(), 6, false, now.Add(2*time.Minute)), 2)
	bronzeB := withTeam(player(uuid.New(), 6, false, now.Add(3*time.Minute)), 2)
	radiant := player(uuid.New(), 25, false, now.Add(4*time.Minute))

	lobby := matchmaking.LobbyPlan{Roster: []matchmaking.Player{plat, iron, bronzeA, bronzeB}}
	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		[]matchmaking.Player{plat, iron, bronzeA, bronzeB, radiant},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.False(t, ids[radiant.UserID], "a Radiant is not a same-window or fringe leftover for this 2v2")
	assert.True(t, ids[plat.UserID])
}

func TestRepairUnfairWindowPairs_OutlierDoesNotTakeReservedCanSub(t *testing.T) {
	now := time.Now()
	plat := withTeam(player(uuid.New(), 13, false, now), 1)
	iron := withTeam(player(uuid.New(), 3, false, now.Add(time.Minute)), 1)
	bronzeA := withTeam(player(uuid.New(), 6, false, now.Add(2*time.Minute)), 2)
	bronzeB := withTeam(player(uuid.New(), 6, false, now.Add(3*time.Minute)), 2)
	goldCS := player(uuid.New(), 11, true, now.Add(4*time.Minute))
	goldCS2 := player(uuid.New(), 11, true, now.Add(5*time.Minute))
	fairA := withTeam(player(uuid.New(), 12, false, now.Add(6*time.Minute)), 1)
	fairB := withTeam(player(uuid.New(), 12, false, now.Add(7*time.Minute)), 1)
	fairC := withTeam(player(uuid.New(), 12, false, now.Add(8*time.Minute)), 2)
	fairD := withTeam(player(uuid.New(), 12, false, now.Add(9*time.Minute)), 2)

	unfair := matchmaking.LobbyPlan{Roster: []matchmaking.Player{plat, iron, bronzeA, bronzeB}}
	fair := matchmaking.LobbyPlan{Roster: []matchmaking.Player{fairA, fairB, fairC, fairD}}
	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{unfair, fair},
		[]matchmaking.Player{plat, iron, bronzeA, bronzeB, goldCS, goldCS2, fairA, fairB, fairC, fairD},
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 2, SubMin: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.False(t, ids[goldCS.UserID])
	assert.False(t, ids[goldCS2.UserID])
	assert.True(t, ids[plat.UserID])
}

func TestRepairTrialAcceptedForTest(t *testing.T) {
	t1, t2 := 1, 2
	sepBefore := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		{AvgRank: 13, TeamNumber: &t1}, {AvgRank: 21, TeamNumber: &t1},
		{AvgRank: 6, TeamNumber: &t2}, {AvgRank: 21, TeamNumber: &t2},
	}}
	sepAfter := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		{AvgRank: 11, TeamNumber: &t1}, {AvgRank: 21, TeamNumber: &t1},
		{AvgRank: 6, TeamNumber: &t2}, {AvgRank: 21, TeamNumber: &t2},
	}}
	assert.True(t, matchmaking.RepairTrialAcceptedForTest(sepBefore, sepAfter, repairSettings(), 25))
	assert.False(t, matchmaking.RepairTrialAcceptedForTest(sepBefore, sepBefore, repairSettings(), 25))

	outBefore := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		{AvgRank: 13, TeamNumber: &t1}, {AvgRank: 3, TeamNumber: &t1},
		{AvgRank: 6, TeamNumber: &t2}, {AvgRank: 6, TeamNumber: &t2},
	}}
	outAfter := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		{AvgRank: 11, TeamNumber: &t1}, {AvgRank: 3, TeamNumber: &t1},
		{AvgRank: 6, TeamNumber: &t2}, {AvgRank: 6, TeamNumber: &t2},
	}}
	assert.True(t, matchmaking.RepairTrialAcceptedForTest(outBefore, outAfter, repairSettings(), 25))

	createsSep := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		{AvgRank: 1, TeamNumber: &t1}, {AvgRank: 3, TeamNumber: &t1},
		{AvgRank: 6, TeamNumber: &t2}, {AvgRank: 6, TeamNumber: &t2},
	}}
	assert.False(t, matchmaking.RepairTrialAcceptedForTest(outBefore, createsSep, repairSettings(), 25))

	fair := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		{AvgRank: 12, TeamNumber: &t1}, {AvgRank: 12, TeamNumber: &t1},
		{AvgRank: 12, TeamNumber: &t2}, {AvgRank: 12, TeamNumber: &t2},
	}}
	assert.False(t, matchmaking.RepairTrialAcceptedForTest(fair, outAfter, repairSettings(), 25))

	sepNoOutlier := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		{AvgRank: 15, TeamNumber: &t1}, {AvgRank: 14, TeamNumber: &t1},
		{AvgRank: 8, TeamNumber: &t2}, {AvgRank: 7, TeamNumber: &t2},
	}}
	createsOutlier := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		{AvgRank: 15, TeamNumber: &t1}, {AvgRank: 14, TeamNumber: &t1},
		{AvgRank: 25, TeamNumber: &t2}, {AvgRank: 7, TeamNumber: &t2},
	}}
	assert.False(t, matchmaking.RepairTrialAcceptedForTest(sepNoOutlier, createsOutlier, repairSettings(), 25))
}

func TestInPlaceTrialNotBetterForTest(t *testing.T) {
	now := time.Now()
	low := player(uuid.New(), 6, false, now)
	high := player(uuid.New(), 13, false, now.Add(time.Minute))
	inEarly := player(uuid.New(), 11, false, now.Add(2*time.Minute))
	inLate := player(uuid.New(), 11, false, now.Add(3*time.Minute))

	assert.True(t, matchmaking.InPlaceTrialNotBetterForTest(true, 1, 0, 2, 0, 1, 1, low, inEarly, low, inEarly),
		"worse team sep is not better")
	assert.False(t, matchmaking.InPlaceTrialNotBetterForTest(true, 2, 0, 1, 0, 1, 1, low, inEarly, low, inEarly),
		"better team sep wins")
	assert.True(t, matchmaking.InPlaceTrialNotBetterForTest(true, 1, 0, 1, 0, 5, 3, low, inEarly, low, inEarly),
		"same sep with a larger rank delta loses")
	assert.False(t, matchmaking.InPlaceTrialNotBetterForTest(true, 1, 0, 1, 0, 2, 3, low, inEarly, low, inEarly),
		"same sep with a smaller rank delta wins")
	assert.True(t, matchmaking.InPlaceTrialNotBetterForTest(true, 1, 0, 1, 0, 3, 3, high, inEarly, low, inEarly),
		"same sep and delta keeps the lower-ranked outgoing seat")
	assert.False(t, matchmaking.InPlaceTrialNotBetterForTest(true, 1, 0, 1, 0, 3, 3, low, inEarly, high, inEarly),
		"same sep and delta prefers replacing the lower-ranked seat")
	assert.True(t, matchmaking.InPlaceTrialNotBetterForTest(true, 1, 0, 1, 0, 3, 3, low, inLate, low, inEarly),
		"same sep, delta, and outgoing keeps the earlier leftover")
	assert.False(t, matchmaking.InPlaceTrialNotBetterForTest(true, 1, 0, 1, 0, 3, 3, low, inEarly, low, inLate),
		"same sep, delta, and outgoing prefers the earlier leftover")

	assert.True(t, matchmaking.InPlaceTrialNotBetterForTest(false, 0, 2, 0, 3, 1, 1, low, inEarly, low, inEarly),
		"worse outlier gap is not better")
	assert.False(t, matchmaking.InPlaceTrialNotBetterForTest(false, 0, 3, 0, 2, 1, 1, low, inEarly, low, inEarly),
		"better outlier gap wins")
	assert.True(t, matchmaking.InPlaceTrialNotBetterForTest(false, 0, 2, 0, 2, 4, 2, low, inEarly, low, inEarly),
		"same outlier gap with a larger rank delta loses")
}

func TestTryRepairUnfairLobbyWindowComboForTest(t *testing.T) {
	now := time.Now()
	cfg := matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"}
	ascA := player(uuid.New(), 21, false, now)
	ascB := player(uuid.New(), 21, false, now.Add(time.Minute))
	plat := player(uuid.New(), 13, false, now.Add(2*time.Minute))
	bronze := player(uuid.New(), 6, false, now.Add(3*time.Minute))
	ironA := player(uuid.New(), 3, false, now.Add(4*time.Minute))
	ironB := player(uuid.New(), 3, false, now.Add(5*time.Minute))
	gold := player(uuid.New(), 11, false, now.Add(6*time.Minute))
	radiant := player(uuid.New(), 25, false, now.Add(7*time.Minute))
	ironCS := player(uuid.New(), 3, true, now.Add(8*time.Minute))

	lobby := windowedLobby(t, []matchmaking.Player{ascA, ascB, plat, bronze}, 2, 25)
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))

	unchanged, ok := matchmaking.TryRepairUnfairLobbyWindowComboForTest(lobby, nil, 0, 0, cfg, repairSettings())
	assert.False(t, ok)
	assert.True(t, rosterIDSet(unchanged)[plat.UserID])

	highOnly, ok := matchmaking.TryRepairUnfairLobbyWindowComboForTest(
		lobby, []matchmaking.Player{radiant}, 0, 0, cfg, repairSettings(),
	)
	assert.False(t, ok, "Radiant combo widens the high window")
	assert.True(t, rosterIDSet(highOnly)[plat.UserID])

	blocked, ok := matchmaking.TryRepairUnfairLobbyWindowComboForTest(
		lobby, []matchmaking.Player{ironCS}, 1, 2, cfg, repairSettings(),
	)
	assert.False(t, ok, "reserved can-subs cannot join via combo")
	assert.True(t, rosterIDSet(blocked)[plat.UserID])

	fixed, ok := matchmaking.TryRepairUnfairLobbyWindowComboForTest(
		lobby, []matchmaking.Player{ironA, ironB, gold}, 0, 0, cfg, repairSettings(),
	)
	assert.True(t, ok)
	ids := rosterIDSet(fixed)
	assert.True(t, ids[ironA.UserID])
	assert.True(t, ids[ironB.UserID])
	assert.False(t, ids[plat.UserID])
	assert.False(t, ids[gold.UserID], "Gold+Iron widens the low window vs two Irons")

	platOut := withTeam(player(uuid.New(), 13, false, now), 1)
	ironOut := withTeam(player(uuid.New(), 3, false, now.Add(time.Minute)), 1)
	bronzeA := withTeam(player(uuid.New(), 6, false, now.Add(2*time.Minute)), 2)
	bronzeB := withTeam(player(uuid.New(), 6, false, now.Add(3*time.Minute)), 2)
	gold2 := player(uuid.New(), 11, false, now.Add(4*time.Minute))
	outlierLobby := matchmaking.LobbyPlan{Roster: []matchmaking.Player{platOut, ironOut, bronzeA, bronzeB}}
	require.Greater(t, matchmaking.OutlierGapForTest(outlierLobby.Roster), 6.0)
	outlierFixed, ok := matchmaking.TryRepairUnfairLobbyWindowComboForTest(
		outlierLobby, []matchmaking.Player{gold2}, 0, 0, cfg, repairSettings(),
	)
	assert.True(t, ok)
	assert.True(t, rosterIDSet(outlierFixed)[gold2.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(outlierFixed, repairSettings(), 25))
}
