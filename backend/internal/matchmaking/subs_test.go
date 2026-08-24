package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func player(id uuid.UUID, rank float64, canSub bool, at time.Time) matchmaking.Player {
	return matchmaking.Player{
		UserID:        id,
		AvgRank:       rank,
		CanSubstitute: canSub,
		CreatedAt:     at,
	}
}

func TestApplySubCapacityRosterConstraint_SkipsSingleLobby(t *testing.T) {
	lobbies := []matchmaking.LobbyPlan{{Roster: []matchmaking.Player{{UserID: uuid.New(), CanSubstitute: true}}}}
	out, adjusted := matchmaking.ApplySubCapacityRosterConstraint(lobbies, nil, matchmaking.Config{
		LobbyCount: 1,
		SubMin:     1,
		TeamSize:   2,
		TierCount:  25,
	})
	assert.True(t, out[0].Roster[0].CanSubstitute)
	assert.False(t, adjusted)
}

func TestApplySubCapacityRosterConstraint_SwapsSubEligibleForNonSub(t *testing.T) {
	now := time.Now()
	subA := uuid.New()
	subB := uuid.New()
	nonSub := uuid.New()

	lobbies := []matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{
			player(subA, 10, true, now),
			player(subB, 9, true, now.Add(time.Minute)),
		}},
		{Roster: []matchmaking.Player{
			player(uuid.New(), 8, true, now.Add(2*time.Minute)),
			player(uuid.New(), 7, true, now.Add(3*time.Minute)),
		}},
	}
	all := []matchmaking.Player{
		player(subA, 10, true, now),
		player(subB, 9, true, now.Add(time.Minute)),
		player(uuid.New(), 8, true, now.Add(2*time.Minute)),
		player(uuid.New(), 7, true, now.Add(3*time.Minute)),
		player(nonSub, 6, false, now.Add(4*time.Minute)),
		player(uuid.New(), 5, false, now.Add(5*time.Minute)),
	}

	// 4 sub-eligible on rosters, 2 lobbies, sub_min=2 → max 0 sub-eligible on rosters; swaps in non-subs.
	out, adjusted := matchmaking.ApplySubCapacityRosterConstraint(lobbies, all, matchmaking.Config{
		LobbyCount: 2,
		SubMin:     2,
		TeamSize:   2,
		TierCount:  25,
	})
	rosteredSubs := 0
	foundNonSub := false
	for _, lobby := range out {
		for _, p := range lobby.Roster {
			if p.CanSubstitute {
				rosteredSubs++
			}
			if p.UserID == nonSub {
				foundNonSub = true
			}
		}
	}
	assert.True(t, foundNonSub)
	assert.True(t, adjusted)
	assert.Less(t, rosteredSubs, 4)
}

func TestAssignMandatorySubs_PlacesPerLobby(t *testing.T) {
	now := time.Now()
	sub1 := player(uuid.New(), 5, true, now)
	sub2 := player(uuid.New(), 4, true, now.Add(time.Minute))
	all := []matchmaking.Player{sub1, sub2}

	lobbies := []matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{}},
		{Roster: []matchmaking.Player{}},
	}

	out := matchmaking.AssignMandatorySubs(lobbies, all, 2, 1)
	assert.Len(t, out[0].Subs, 1)
	assert.Len(t, out[1].Subs, 1)
	assert.Nil(t, out[0].Subs[0].TeamNumber)
}

func TestSwapRosterPlayerForTest(t *testing.T) {
	removeID := uuid.New()
	replacement := player(uuid.New(), 9, false, time.Now())
	lobbies := []matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{player(removeID, 10, false, time.Now())}},
	}

	out := matchmaking.SwapRosterPlayerForTest(lobbies, removeID, replacement)
	assert.Equal(t, replacement.UserID, out[0].Roster[0].UserID)

	lobbies2 := []matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{player(removeID, 10, false, time.Now())}},
	}
	unchanged := matchmaking.SwapRosterPlayerForTest(lobbies2, uuid.New(), replacement)
	assert.Equal(t, removeID, unchanged[0].Roster[0].UserID)
}

func TestAssignRemainingAsSubs_RoundRobin(t *testing.T) {
	now := time.Now()
	sub1 := player(uuid.New(), 5, true, now)
	sub2 := player(uuid.New(), 4, true, now.Add(time.Minute))
	sub3 := player(uuid.New(), 3, true, now.Add(2*time.Minute))

	lobbies := []matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{}},
		{Roster: []matchmaking.Player{}},
	}

	out := matchmaking.AssignRemainingAsSubs(lobbies, []matchmaking.Player{sub1, sub2, sub3})
	assert.Len(t, out[0].Subs, 2)
	assert.Len(t, out[1].Subs, 1)
}

func TestApplySubCapacityRosterConstraint_DoesNotInsertRadiantIntoLowWindow(t *testing.T) {
	now := time.Now()
	silver := uuid.New()
	radiant := uuid.New()
	iron := uuid.New()
	goldA := uuid.New()
	goldB := uuid.New()

	lobbies := []matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{
			player(uuid.New(), 21, false, now),
			player(uuid.New(), 21, false, now.Add(time.Minute)),
			player(silver, 8, true, now.Add(2*time.Minute)),
			player(uuid.New(), 6, false, now.Add(3*time.Minute)),
		}},
		{Roster: []matchmaking.Player{
			player(uuid.New(), 21, false, now.Add(4*time.Minute)),
			player(uuid.New(), 21, false, now.Add(5*time.Minute)),
			player(goldA, 11, true, now.Add(6*time.Minute)),
			player(goldB, 11, true, now.Add(7*time.Minute)),
		}},
	}
	all := append([]matchmaking.Player{}, lobbies[0].Roster...)
	all = append(all, lobbies[1].Roster...)
	all = append(all,
		player(radiant, 25, false, now.Add(8*time.Minute)),
		player(iron, 3, false, now.Add(9*time.Minute)),
		player(uuid.New(), 11, true, now.Add(10*time.Minute)),
		player(uuid.New(), 13, true, now.Add(11*time.Minute)),
		player(uuid.New(), 8, true, now.Add(12*time.Minute)),
	)

	out, adjusted := matchmaking.ApplySubCapacityRosterConstraint(allLobbiesCopy(lobbies), all, matchmaking.Config{
		LobbyCount: 2,
		SubMin:     2,
		TeamSize:   2,
		TierCount:  25,
	})
	assert.True(t, adjusted)

	foundIron, foundRadiant, foundSilver := false, false, false
	for _, lobby := range out {
		for _, p := range lobby.Roster {
			if p.UserID == iron {
				foundIron = true
			}
			if p.UserID == radiant {
				foundRadiant = true
			}
			if p.UserID == silver {
				foundSilver = true
			}
		}
	}
	assert.True(t, foundIron, "low-window Iron should fill the vacated Silver slot")
	assert.False(t, foundRadiant, "Radiant must not be pulled into a low-window hole")
	assert.False(t, foundSilver, "lowest rostered can-sub should be swapped off")
}

func allLobbiesCopy(lobbies []matchmaking.LobbyPlan) []matchmaking.LobbyPlan {
	out := make([]matchmaking.LobbyPlan, len(lobbies))
	for i, lobby := range lobbies {
		out[i].Roster = append([]matchmaking.Player(nil), lobby.Roster...)
		out[i].Subs = append([]matchmaking.Player(nil), lobby.Subs...)
	}
	return out
}

func TestBestNonSubReplacement_PrefersSameWindowOverHigherRank(t *testing.T) {
	now := time.Now()
	vacated := player(uuid.New(), 12, true, now)
	plat := player(uuid.New(), 14, false, now.Add(time.Minute))
	radiant := player(uuid.New(), 25, false, now.Add(2*time.Minute))
	iron := player(uuid.New(), 3, false, now.Add(3*time.Minute))
	lobbies := []matchmaking.LobbyPlan{{Roster: []matchmaking.Player{vacated}}}

	got, ok := matchmaking.BestNonSubReplacementForTest(
		[]matchmaking.Player{vacated, plat, radiant, iron},
		lobbies,
		vacated,
		5,
		25,
	)
	require.True(t, ok)
	assert.Equal(t, plat.UserID, got.UserID)
}

func TestBestNonSubReplacement_FallsBackToNearestWindow(t *testing.T) {
	now := time.Now()
	vacated := player(uuid.New(), 12, true, now)
	silver := player(uuid.New(), 8, false, now.Add(time.Minute))
	radiant := player(uuid.New(), 25, false, now.Add(2*time.Minute))
	lobbies := []matchmaking.LobbyPlan{{Roster: []matchmaking.Player{vacated}}}

	got, ok := matchmaking.BestNonSubReplacementForTest(
		[]matchmaking.Player{vacated, silver, radiant},
		lobbies,
		vacated,
		5,
		25,
	)
	require.True(t, ok)
	assert.Equal(t, silver.UserID, got.UserID)
}

func TestBestNonSubReplacement_NoWindowsUsesClosestRank(t *testing.T) {
	now := time.Now()
	vacated := player(uuid.New(), 12, true, now)
	iron := player(uuid.New(), 3, false, now.Add(time.Minute))
	radiant := player(uuid.New(), 25, false, now.Add(2*time.Minute))
	lobbies := []matchmaking.LobbyPlan{{Roster: []matchmaking.Player{vacated}}}

	got, ok := matchmaking.BestNonSubReplacementForTest(
		[]matchmaking.Player{vacated, iron, radiant},
		lobbies,
		vacated,
		0,
		0,
	)
	require.True(t, ok)
	assert.Equal(t, iron.UserID, got.UserID)
}

func TestBestNonSubReplacement_NoCandidates(t *testing.T) {
	now := time.Now()
	vacated := player(uuid.New(), 12, true, now)
	_, ok := matchmaking.BestNonSubReplacementForTest(
		[]matchmaking.Player{vacated, player(uuid.New(), 8, true, now.Add(time.Minute))},
		[]matchmaking.LobbyPlan{{Roster: []matchmaking.Player{vacated}}},
		vacated,
		2,
		25,
	)
	assert.False(t, ok)
}

func TestPickCloserToTarget(t *testing.T) {
	now := time.Now()
	near := player(uuid.New(), 8, false, now)
	far := player(uuid.New(), 25, false, now.Add(time.Minute))
	got := matchmaking.PickCloserToTargetForTest([]matchmaking.Player{far, near}, 7)
	assert.Equal(t, near.UserID, got.UserID)
	empty := matchmaking.PickCloserToTargetForTest(nil, 7)
	assert.Equal(t, uuid.Nil, empty.UserID)
}
