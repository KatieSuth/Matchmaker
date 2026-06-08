package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
	out := matchmaking.ApplySubCapacityRosterConstraint(lobbies, nil, 1, 1)
	assert.True(t, out[0].Roster[0].CanSubstitute)
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
	out := matchmaking.ApplySubCapacityRosterConstraint(lobbies, all, 2, 2)
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
