package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seated builds a rostered player on a team for direct repair-stage tests.
func seated(rank float64, team int, at time.Time) matchmaking.Player {
	return withTeam(player(uuid.New(), rank, false, at), team)
}

func outlierCfg(teamSize int) matchmaking.Config {
	return matchmaking.Config{TeamSize: teamSize, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"}
}

// The reported case: a lone 19 (Diamond 3) is 7 above the next player. Same-window, fringe,
// and combo repair cannot use the leftover 21 (no one is seated in its band), so only the
// held-out 13 used to qualify. A 21 on the other team fixes it without them.
func TestPlanEvent_OutlierPartnerAvoidsDeprioritizedPlayer(t *testing.T) {
	now := time.Now()
	ranks := []float64{13, 9, 9, 19, 21, 7, 11, 12, 8, 9, 12, 8}
	var players []matchmaking.Player
	for i, r := range ranks {
		p := player(uuid.New(), r, false, now.Add(time.Duration(i)*time.Second))
		if i == 0 {
			p.NoShowCount = 4
		}
		players = append(players, p)
	}

	plan, err := matchmaking.PlanEvent(players, fiveVFiveCfg(3), repairSettings())
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)
	roster, _ := planRosterAndSubIDs(plan)
	assert.False(t, roster[players[0].UserID], "the held-out player is not needed")
	assert.True(t, roster[players[4].UserID], "the 21 partners the lone 19")
	assert.False(t, plan.Lobbies[0].FairnessWarning)

	team1, team2 := matchmaking.SplitRosterByTeamNumberForTest(plan.Lobbies[0].Roster)
	var top19Team, top21Team int
	for _, p := range plan.Lobbies[0].Roster {
		switch p.AvgRank {
		case 19:
			top19Team = *p.TeamNumber
		case 21:
			top21Team = *p.TeamNumber
		}
	}
	assert.NotEqual(t, top19Team, top21Team, "the partner sits on the other team")
	assert.Less(t, matchmaking.TeamAverageSeparationForTest(team1, team2), 3.0)
}

// Without deprioritization the same lone-outlier shape used to stay flagged; now it is fixed.
func TestRepair_OutlierPartnerFixesLobbyOtherStagesCannot(t *testing.T) {
	now := time.Now()
	lobby := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		seated(9, 1, now), seated(9, 1, now), seated(7, 1, now), seated(12, 1, now), seated(19, 1, now),
		seated(9, 2, now), seated(8, 2, now), seated(8, 2, now), seated(12, 2, now), seated(11, 2, now),
	}}
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))
	partner := player(uuid.New(), 21, false, now)

	all := append(append([]matchmaking.Player{}, lobby.Roster...), partner)
	out := matchmaking.RepairUnfairWindowPairsForTest([]matchmaking.LobbyPlan{lobby}, all, outlierCfg(5), repairSettings())
	assert.True(t, rosterIDSet(out[0])[partner.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], repairSettings(), 25))
}

func TestTryOutlierPartnerSwap(t *testing.T) {
	now := time.Now()
	cfg := outlierCfg(2)
	// 20 is 8 above the 12; teams 16 vs 5.
	lopsided := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		seated(20, 1, now), seated(12, 1, now), seated(6, 2, now), seated(4, 2, now),
	}}
	// 20 is 10 above the 10; teams 15 vs 9.
	fixable := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
		seated(20, 1, now), seated(10, 1, now), seated(10, 2, now), seated(8, 2, now),
	}}

	t.Run("seats a near-rank leftover on the other team", func(t *testing.T) {
		partner := player(uuid.New(), 18, false, now)
		out, ok := matchmaking.TryOutlierPartnerSwapForTest(fixable, []matchmaking.Player{partner}, 0, 0, cfg, repairSettings())
		require.True(t, ok)
		for _, p := range out.Roster {
			if p.UserID == partner.UserID {
				assert.Equal(t, 2, *p.TeamNumber)
			}
		}
		assert.False(t, matchmaking.IsLobbyUnfair(out, repairSettings(), 25))
	})

	t.Run("rejects a partner that leaves the lobby unfair", func(t *testing.T) {
		out, ok := matchmaking.TryOutlierPartnerSwapForTest(lopsided, []matchmaking.Player{player(uuid.New(), 14, false, now)}, 0, 0, cfg, repairSettings())
		assert.False(t, ok)
		assert.Equal(t, lopsided.Roster, out.Roster)
	})

	t.Run("skips leftovers beyond the outlier gap", func(t *testing.T) {
		_, ok := matchmaking.TryOutlierPartnerSwapForTest(fixable, []matchmaking.Player{player(uuid.New(), 13, false, now)}, 0, 0, cfg, repairSettings())
		assert.False(t, ok)
	})

	t.Run("only runs when the outlier check fails", func(t *testing.T) {
		sepOnly := matchmaking.LobbyPlan{Roster: []matchmaking.Player{
			seated(21, 1, now), seated(21, 2, now), seated(13, 1, now), seated(6, 2, now),
		}}
		_, ok := matchmaking.TryOutlierPartnerSwapForTest(sepOnly, []matchmaking.Player{player(uuid.New(), 18, false, now)}, 0, 0, cfg, repairSettings())
		assert.False(t, ok)
	})

	t.Run("never takes a reserved can-sub", func(t *testing.T) {
		canSub := player(uuid.New(), 18, true, now)
		_, ok := matchmaking.TryOutlierPartnerSwapForTest(fixable, []matchmaking.Player{canSub}, 1, 1, cfg, repairSettings())
		assert.False(t, ok)
	})

	t.Run("prefers a clean partner over a deprioritized one that balances better", func(t *testing.T) {
		clean := player(uuid.New(), 17, false, now)
		held := player(uuid.New(), 18, false, now)
		held.Deprioritized = true
		held.NoShowCount = 5
		out, ok := matchmaking.TryOutlierPartnerSwapForTest(fixable, []matchmaking.Player{held, clean}, 0, 0, cfg, repairSettings())
		require.True(t, ok)
		ids := rosterIDSet(out)
		assert.True(t, ids[clean.UserID])
		assert.False(t, ids[held.UserID])
	})

	t.Run("uses a deprioritized partner when no clean one works", func(t *testing.T) {
		held := player(uuid.New(), 18, false, now)
		held.Deprioritized = true
		out, ok := matchmaking.TryOutlierPartnerSwapForTest(fixable, []matchmaking.Player{held}, 0, 0, cfg, repairSettings())
		require.True(t, ok)
		assert.True(t, rosterIDSet(out)[held.UserID])
	})

	t.Run("no seated players", func(t *testing.T) {
		subsOnly := matchmaking.LobbyPlan{Roster: []matchmaking.Player{player(uuid.New(), 20, false, now), player(uuid.New(), 5, false, now)}}
		_, ok := matchmaking.TryOutlierPartnerSwapForTest(subsOnly, []matchmaking.Player{player(uuid.New(), 18, false, now)}, 0, 0, cfg, repairSettings())
		assert.False(t, ok)
	})
}

func TestTopSeatedPlayer(t *testing.T) {
	now := time.Now()
	sub := player(uuid.New(), 25, false, now)
	high := seated(20, 2, now)
	_, ok := matchmaking.TopSeatedPlayerForTest([]matchmaking.Player{sub})
	assert.False(t, ok)
	got, ok := matchmaking.TopSeatedPlayerForTest([]matchmaking.Player{sub, seated(10, 1, now), high})
	require.True(t, ok)
	assert.Equal(t, high.UserID, got.UserID, "subs are ignored")
}

func TestBringsInDeprioritized(t *testing.T) {
	stay := matchmaking.Player{UserID: uuid.New(), Deprioritized: true}
	joiner := matchmaking.Player{UserID: uuid.New(), Deprioritized: true}
	clean := matchmaking.Player{UserID: uuid.New()}
	assert.False(t, matchmaking.BringsInDeprioritizedForTest([]matchmaking.Player{stay}, []matchmaking.Player{stay, clean}))
	assert.True(t, matchmaking.BringsInDeprioritizedForTest([]matchmaking.Player{stay}, []matchmaking.Player{stay, joiner}))
}

func TestTryRepairUnfairLobbyWindowCombo_DeprioritizedOnlyWhenNoCleanCombo(t *testing.T) {
	now := time.Now()
	cfg := outlierCfg(2)
	lobby := windowedLobby(t, []matchmaking.Player{
		player(uuid.New(), 21, false, now), player(uuid.New(), 21, false, now.Add(time.Minute)),
		player(uuid.New(), 13, false, now.Add(2*time.Minute)), player(uuid.New(), 6, false, now.Add(3*time.Minute)),
	}, 2, 25)
	held := func(at time.Duration) matchmaking.Player {
		p := player(uuid.New(), 3, false, now.Add(at))
		p.Deprioritized = true
		p.NoShowCount = 4
		return p
	}

	// Only held-out Irons can fill the low window: the combo uses them.
	heldA, heldB := held(4*time.Minute), held(5*time.Minute)
	out, ok := matchmaking.TryRepairUnfairLobbyWindowComboForTest(lobby, []matchmaking.Player{heldA, heldB}, 0, 0, cfg, repairSettings())
	require.True(t, ok)
	ids := rosterIDSet(out)
	assert.True(t, ids[heldA.UserID] || ids[heldB.UserID])

	// A clean Iron that also makes the lobby fair wins over the held-out ones.
	clean := player(uuid.New(), 3, false, now.Add(6*time.Minute))
	out, ok = matchmaking.TryRepairUnfairLobbyWindowComboForTest(lobby, []matchmaking.Player{heldA, heldB, clean}, 0, 0, cfg, repairSettings())
	require.True(t, ok)
	assert.True(t, rosterIDSet(out)[clean.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(out, repairSettings(), 25))
}

func TestMoreBalanced(t *testing.T) {
	// A fair result beats an unfair one even with a bigger team gap.
	assert.True(t, matchmaking.MoreBalancedForTest(2.9, 5, false, 3, 0.1, 0, true, 0))
	assert.False(t, matchmaking.MoreBalancedForTest(0.1, 0, true, 0, 2.9, 5, false, 3))
	// Both fair: the smaller team-average gap wins regardless of stage.
	assert.True(t, matchmaking.MoreBalancedForTest(0.8, 4, false, 3, 2.8, 1, false, 0))
	// Same gap: the smaller outlier gap wins.
	assert.True(t, matchmaking.MoreBalancedForTest(1, 2, false, 2, 1, 4, false, 0))
	// Full tie: the less disruptive (earlier) stage wins.
	assert.True(t, matchmaking.MoreBalancedForTest(1, 2, false, 0, 1, 2, false, 2))
	assert.False(t, matchmaking.MoreBalancedForTest(1, 2, false, 2, 1, 2, false, 0))
}

// Across steps: the most balanced fix brings in the held-out 11, but the clean 14's fix from a
// different step also leaves the lobby fair, so the clean fix is used.
func TestRepair_CleanFixFromAnotherStepBeatsMoreBalancedHeldOutFix(t *testing.T) {
	now := time.Now()
	roster := []matchmaking.Player{seated(10, 1, now), seated(1, 1, now), seated(3, 2, now), seated(16, 2, now)}
	lobby := matchmaking.LobbyPlan{Roster: roster}
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))
	held := heldOut(player(uuid.New(), 11, false, now.Add(time.Minute)), 4)
	clean := player(uuid.New(), 14, false, now.Add(2*time.Minute))

	// Sanity: if the 11 were not held out, their fix would be the one chosen.
	notHeld := held
	notHeld.Deprioritized = false
	plain := matchmaking.RepairUnfairWindowPairsForTest([]matchmaking.LobbyPlan{lobby}, append(append([]matchmaking.Player{}, roster...), notHeld, clean), outlierCfg(2), repairSettings())
	require.True(t, rosterIDSet(plain[0])[notHeld.UserID])

	out := matchmaking.RepairUnfairWindowPairsForTest([]matchmaking.LobbyPlan{lobby}, append(append([]matchmaking.Player{}, roster...), held, clean), outlierCfg(2), repairSettings())
	ids := rosterIDSet(out[0])
	assert.True(t, ids[clean.UserID])
	assert.False(t, ids[held.UserID], "a fair clean fix beats the held-out player")
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], repairSettings(), 25))
}
