package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rankedCfg() matchmaking.Config {
	return matchmaking.Config{TeamSize: 2, Slots: 4, TierCount: 25, LobbyCount: 1, SortLogic: "ranked"}
}

// rankedLobby snake-drafts roster the way ranked mode does.
func rankedLobby(roster ...matchmaking.Player) matchmaking.LobbyPlan {
	team1, team2 := matchmaking.SplitIntoTeams(roster, 2)
	return matchmaking.LobbyPlan{Roster: append(team1, team2...)}
}

func heldOut(p matchmaking.Player, noShows int) matchmaking.Player {
	p.Deprioritized = true
	p.NoShowCount = noShows
	return p
}

// A lone 17 over 10s: the held-out 15 replacing the 17 is the only fair lineup.
func outlierRankedLobby(now time.Time) (matchmaking.LobbyPlan, []matchmaking.Player) {
	roster := []matchmaking.Player{
		player(uuid.New(), 17, false, now), player(uuid.New(), 10, false, now.Add(time.Second)),
		player(uuid.New(), 10, false, now.Add(2*time.Second)), player(uuid.New(), 9, false, now.Add(3*time.Second)),
	}
	return rankedLobby(roster...), roster
}

func TestRepairUnfairRankedLobbies_SeatsHeldOutPlayerWhenOnlyTheyFixIt(t *testing.T) {
	now := time.Now()
	lobby, roster := outlierRankedLobby(now)
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))
	held := heldOut(player(uuid.New(), 15, false, now.Add(time.Minute)), 4)

	out := matchmaking.RepairUnfairRankedLobbiesForTest([]matchmaking.LobbyPlan{lobby}, append(append([]matchmaking.Player{}, roster...), held), rankedCfg(), repairSettings())
	assert.True(t, rosterIDSet(out[0])[held.UserID], "the 15 gives the lone 17 a near-rank opponent")
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], repairSettings(), 25))
}

func TestRepairUnfairRankedLobbies_CleanLeftoverWinsWhenEquallyBalanced(t *testing.T) {
	now := time.Now()
	lobby, roster := outlierRankedLobby(now)
	held := heldOut(player(uuid.New(), 15, false, now.Add(-time.Hour)), 4)
	clean := player(uuid.New(), 15, false, now.Add(time.Minute))

	all := append(append([]matchmaking.Player{}, roster...), held, clean)
	out := matchmaking.RepairUnfairRankedLobbiesForTest([]matchmaking.LobbyPlan{lobby}, all, rankedCfg(), repairSettings())
	ids := rosterIDSet(out[0])
	assert.True(t, ids[clean.UserID], "same rank, same balance: the clean player wins the tie")
	assert.False(t, ids[held.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], repairSettings(), 25))
}

func TestRepairUnfairRankedLobbies_MoreBalancedHeldOutPlayerBeatsCleanLeftover(t *testing.T) {
	now := time.Now()
	lobby, roster := outlierRankedLobby(now)
	held := heldOut(player(uuid.New(), 15, false, now.Add(time.Minute)), 4)
	// The 14 also makes the lobby fair, just less evenly than the 15.
	clean := player(uuid.New(), 14, false, now.Add(-time.Hour))

	all := append(append([]matchmaking.Player{}, roster...), held, clean)
	out := matchmaking.RepairUnfairRankedLobbiesForTest([]matchmaking.LobbyPlan{lobby}, all, rankedCfg(), repairSettings())
	ids := rosterIDSet(out[0])
	assert.True(t, ids[held.UserID], "best balance wins")
	assert.False(t, ids[clean.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], repairSettings(), 25))
}

func TestRepairUnfairRankedLobbies_NoHeldOutPlayerKeepsRankedBand(t *testing.T) {
	now := time.Now()
	lobby, roster := outlierRankedLobby(now)
	clean := player(uuid.New(), 14, false, now.Add(time.Minute))

	all := append(append([]matchmaking.Player{}, roster...), clean)
	out := matchmaking.RepairUnfairRankedLobbiesForTest([]matchmaking.LobbyPlan{lobby}, all, rankedCfg(), repairSettings())
	assert.Equal(t, lobby.Roster, out[0].Roster, "ranked never swaps clean leftovers in on its own")
}

func TestRepairUnfairRankedLobbies_HeldOutPlayerWhoCannotFixItStaysOut(t *testing.T) {
	now := time.Now()
	lobby, roster := outlierRankedLobby(now)
	held := heldOut(player(uuid.New(), 25, false, now.Add(time.Minute)), 4)
	clean := player(uuid.New(), 14, false, now.Add(2*time.Minute))

	all := append(append([]matchmaking.Player{}, roster...), held, clean)
	out := matchmaking.RepairUnfairRankedLobbiesForTest([]matchmaking.LobbyPlan{lobby}, all, rankedCfg(), repairSettings())
	assert.Equal(t, lobby.Roster, out[0].Roster, "no fair swap involves the held-out player, so nothing changes")
}

func TestRepairUnfairRankedLobbies_FairLobbyUntouched(t *testing.T) {
	now := time.Now()
	roster := []matchmaking.Player{
		player(uuid.New(), 10, false, now), player(uuid.New(), 10, false, now),
		player(uuid.New(), 9, false, now), player(uuid.New(), 9, false, now),
	}
	lobby := rankedLobby(roster...)
	held := heldOut(player(uuid.New(), 10, false, now), 4)
	out := matchmaking.RepairUnfairRankedLobbiesForTest([]matchmaking.LobbyPlan{lobby}, append(roster, held), rankedCfg(), repairSettings())
	assert.Equal(t, lobby.Roster, out[0].Roster)
}

func TestRepairUnfairRankedLobbies_SkipsReservedCanSubs(t *testing.T) {
	now := time.Now()
	lobby, roster := outlierRankedLobby(now)
	held := heldOut(player(uuid.New(), 15, true, now.Add(time.Minute)), 4)
	cfg := rankedCfg()
	cfg.LobbyCount, cfg.SubMin = 2, 1 // two reserved can-subs; only one exists

	other := rankedLobby(
		player(uuid.New(), 5, false, now), player(uuid.New(), 5, false, now),
		player(uuid.New(), 5, false, now), player(uuid.New(), 5, false, now),
	)
	all := append(append(append([]matchmaking.Player{}, roster...), other.Roster...), held)
	out := matchmaking.RepairUnfairRankedLobbiesForTest([]matchmaking.LobbyPlan{lobby, other}, all, cfg, repairSettings())
	assert.False(t, rosterIDSet(out[0])[held.UserID])
}

func TestRepairUnfairRankedLobbies_RedraftsWhenKeepingSeatsIsNotEnough(t *testing.T) {
	now := time.Now()
	// 18 and 17 on one team after the snake draft; a held-out 16 in for a 9 only balances
	// once the lobby is re-drafted (18+9 vs 17+16 keeps sides lopsided in place).
	roster := []matchmaking.Player{
		player(uuid.New(), 18, false, now), player(uuid.New(), 8, false, now.Add(time.Second)),
		player(uuid.New(), 9, false, now.Add(2*time.Second)), player(uuid.New(), 7, false, now.Add(3*time.Second)),
	}
	lobby := rankedLobby(roster...)
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))
	held := heldOut(player(uuid.New(), 16, false, now.Add(time.Minute)), 4)

	out := matchmaking.RepairUnfairRankedLobbiesForTest([]matchmaking.LobbyPlan{lobby}, append(append([]matchmaking.Player{}, roster...), held), rankedCfg(), repairSettings())
	assert.True(t, rosterIDSet(out[0])[held.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], repairSettings(), 25))
}

func TestPlanEvent_RankedSeatsHeldOutPlayerNeededForFairness(t *testing.T) {
	now := time.Now()
	ranks := []float64{15, 10, 10, 17, 8, 4, 6, 1, 8, 8, 24, 9}
	var players []matchmaking.Player
	for i, r := range ranks {
		p := player(uuid.New(), r, false, now.Add(time.Duration(i)*time.Second))
		if i == 0 {
			p.NoShowCount = 4
		}
		players = append(players, p)
	}
	cfg := fiveVFiveCfg(3)
	cfg.SortLogic = "ranked"

	plan, err := matchmaking.PlanEvent(players, cfg, repairSettings())
	require.NoError(t, err)
	roster, _ := planRosterAndSubIDs(plan)
	assert.True(t, roster[players[0].UserID])
	assert.False(t, plan.Lobbies[0].FairnessWarning)
}

func TestRepairUnfairRankedLobbies_NeverReplacesAnUnseatedPlayer(t *testing.T) {
	now := time.Now()
	lobby, roster := outlierRankedLobby(now)
	unseated := player(uuid.New(), 2, false, now.Add(time.Second)) // no TeamNumber
	lobby.Roster = append(lobby.Roster, unseated)
	held := heldOut(player(uuid.New(), 15, false, now.Add(time.Minute)), 4)

	all := append(append([]matchmaking.Player{}, roster...), unseated, held)
	out := matchmaking.RepairUnfairRankedLobbiesForTest([]matchmaking.LobbyPlan{lobby}, all, rankedCfg(), repairSettings())
	ids := rosterIDSet(out[0])
	assert.True(t, ids[held.UserID])
	assert.True(t, ids[unseated.UserID], "only seated players can be swapped out")
}
