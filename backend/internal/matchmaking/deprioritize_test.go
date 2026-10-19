package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noShowPlayer is a registrant with a no-show count, registered `minutes` after base.
func noShowPlayer(rank float64, canSub bool, noShows int, base time.Time, minutes int) matchmaking.Player {
	p := player(uuid.New(), rank, canSub, base.Add(time.Duration(minutes)*time.Minute))
	p.NoShowCount = noShows
	return p
}

// deprioritizedByID maps each player to their Deprioritized flag.
func deprioritizedByID(players []matchmaking.Player) map[uuid.UUID]bool {
	out := make(map[uuid.UUID]bool, len(players))
	for _, p := range players {
		out[p.UserID] = p.Deprioritized
	}
	return out
}

// planRosterAndSubIDs splits a plan's placements into roster and sub ID sets.
func planRosterAndSubIDs(plan matchmaking.GamePlan) (roster, subs map[uuid.UUID]bool) {
	roster, subs = map[uuid.UUID]bool{}, map[uuid.UUID]bool{}
	for _, lobby := range plan.Lobbies {
		for _, p := range lobby.Roster {
			roster[p.UserID] = true
		}
		for _, p := range lobby.Subs {
			subs[p.UserID] = true
		}
	}
	return roster, subs
}

func fiveVFiveCfg(threshold int) matchmaking.Config {
	return matchmaking.Config{
		EventID: uuid.New(), TeamSize: 5, Slots: 10, TierCount: 25, LobbyCount: 1,
		SortLogic: "balanced", GameLabel: "Game 1 (5v5)", NoShowThreshold: threshold,
	}
}

func TestMarkDeprioritized_DisabledIgnoresCounts(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{noShowPlayer(10, false, 9, now, 0)}
	players[0].Deprioritized = true // stale caller value must be cleared

	out := matchmaking.MarkDeprioritizedForTest(players, fiveVFiveCfg(0))
	assert.False(t, out[0].Deprioritized)
}

func TestMarkDeprioritized_HoldsOutWhenEnoughOthers(t *testing.T) {
	now := time.Now()
	var players []matchmaking.Player
	for i := 0; i < 10; i++ {
		players = append(players, noShowPlayer(10, false, i%3, now, i)) // counts 0..2, all below 3
	}
	a := noShowPlayer(10, false, 4, now, 20)
	b := noShowPlayer(10, false, 3, now, 21)
	players = append(players, a, b)

	flags := deprioritizedByID(matchmaking.MarkDeprioritizedForTest(players, fiveVFiveCfg(3)))
	assert.True(t, flags[a.UserID])
	assert.True(t, flags[b.UserID], "count equal to the threshold counts as excessive")
	assert.Len(t, matchmaking.EligiblePlayersForTest(matchmaking.MarkDeprioritizedForTest(players, fiveVFiveCfg(3))), 10)
}

// The user's example: 10 needed, 11 registered, A has 4 and B has 3 with a threshold of 3.
func TestMarkDeprioritized_AddsBackFewestNoShowsFirst(t *testing.T) {
	now := time.Now()
	var players []matchmaking.Player
	for i := 0; i < 9; i++ {
		players = append(players, noShowPlayer(10, false, 0, now, i))
	}
	a := noShowPlayer(10, false, 4, now, 10)
	b := noShowPlayer(10, false, 3, now, 11)
	players = append(players, a, b)

	flags := deprioritizedByID(matchmaking.MarkDeprioritizedForTest(players, fiveVFiveCfg(3)))
	assert.True(t, flags[a.UserID])
	assert.False(t, flags[b.UserID])
}

func TestMarkDeprioritized_AddBackTieUsesAvailability(t *testing.T) {
	now := time.Now()
	var players []matchmaking.Player
	for i := 0; i < 9; i++ {
		players = append(players, noShowPlayer(10, false, 0, now, i))
	}
	busy := noShowPlayer(10, false, 3, now, 10)
	busy.RegisteredGameCount = 3
	lighter := noShowPlayer(10, false, 3, now, 11)
	lighter.RegisteredGameCount = 1
	players = append(players, busy, lighter)

	flags := deprioritizedByID(matchmaking.MarkDeprioritizedForTest(players, fiveVFiveCfg(3)))
	assert.False(t, flags[lighter.UserID], "fewer group games wins the tie")
	assert.True(t, flags[busy.UserID])
}

func TestMarkDeprioritized_AddsBackNonSubWhenSubCapRequiresIt(t *testing.T) {
	now := time.Now()
	cfg := matchmaking.Config{TeamSize: 2, Slots: 4, TierCount: 25, LobbyCount: 2, SubMin: 1, SortLogic: "balanced", NoShowThreshold: 3}
	var players []matchmaking.Player
	for i := 0; i < 5; i++ {
		players = append(players, noShowPlayer(10, false, 0, now, i))
	}
	for i := 0; i < 3; i++ {
		players = append(players, noShowPlayer(10, true, 0, now, 10+i))
	}
	nonSub := noShowPlayer(10, false, 5, now, 20)
	canSub := noShowPlayer(10, true, 3, now, 21)
	players = append(players, nonSub, canSub)

	// 8 seats, 4 can-subs with 2 reserved → at most 2 can-subs rostered, so 6 non-subs are needed.
	flags := deprioritizedByID(matchmaking.MarkDeprioritizedForTest(players, cfg))
	assert.False(t, flags[nonSub.UserID], "the only held-out non-sub is needed despite more no-shows")
	assert.True(t, flags[canSub.UserID])
}

func TestMarkDeprioritized_RecomputedPerLobbyCount(t *testing.T) {
	now := time.Now()
	cfg := matchmaking.Config{TeamSize: 2, Slots: 4, TierCount: 25, SortLogic: "balanced", NoShowThreshold: 2}
	var players []matchmaking.Player
	for i := 0; i < 6; i++ {
		players = append(players, noShowPlayer(10, false, 0, now, i))
	}
	excessive := noShowPlayer(10, false, 2, now, 10)
	players = append(players, excessive)

	cfg.LobbyCount = 1
	assert.True(t, deprioritizedByID(matchmaking.MarkDeprioritizedForTest(players, cfg))[excessive.UserID])
	cfg.LobbyCount = 2
	assert.False(t, deprioritizedByID(matchmaking.MarkDeprioritizedForTest(players, cfg))[excessive.UserID])
}

func TestMarkDeprioritized_DoesNotMutateInput(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{noShowPlayer(10, false, 5, now, 0)}
	_ = matchmaking.MarkDeprioritizedForTest(players, fiveVFiveCfg(1))
	assert.False(t, players[0].Deprioritized)
}

func TestPlanEvent_DeprioritizationLeavesExcessivePlayerOut(t *testing.T) {
	now := time.Now()
	var players []matchmaking.Player
	for i := 0; i < 9; i++ {
		players = append(players, noShowPlayer(10, false, 0, now, i))
	}
	a := noShowPlayer(10, false, 4, now, 10)
	b := noShowPlayer(10, false, 3, now, 11)
	players = append(players, a, b)

	plan, err := matchmaking.PlanEvent(players, fiveVFiveCfg(3), repairSettings())
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)
	roster, _ := planRosterAndSubIDs(plan)
	assert.False(t, roster[a.UserID])
	assert.True(t, roster[b.UserID])
	for _, p := range plan.Lobbies[0].Roster {
		assert.False(t, p.Deprioritized)
	}
}

func TestPlanEvent_DeprioritizationDisabledMatchesBaseline(t *testing.T) {
	now := time.Now()
	var players []matchmaking.Player
	for i := 0; i < 11; i++ {
		players = append(players, noShowPlayer(float64(5+i), false, 10, now, i))
	}
	cfg := fiveVFiveCfg(0)

	baseline := make([]matchmaking.Player, len(players))
	copy(baseline, players)
	for i := range baseline {
		baseline[i].NoShowCount = 0
	}

	withCounts, err := matchmaking.PlanEvent(players, cfg, repairSettings())
	require.NoError(t, err)
	without, err := matchmaking.PlanEvent(baseline, cfg, repairSettings())
	require.NoError(t, err)

	gotRoster, _ := planRosterAndSubIDs(withCounts)
	wantRoster, _ := planRosterAndSubIDs(without)
	assert.Equal(t, wantRoster, gotRoster)
}

func TestPlanEvent_DeprioritizationKeepsLobbyCount(t *testing.T) {
	now := time.Now()
	cfg := matchmaking.Config{EventID: uuid.New(), TeamSize: 2, Slots: 4, TierCount: 25, SortLogic: "balanced", GameLabel: "Game 1 (2v2)", NoShowThreshold: 1}
	var players []matchmaking.Player
	for i := 0; i < 6; i++ {
		players = append(players, noShowPlayer(10, false, 0, now, i))
	}
	x := noShowPlayer(10, false, 2, now, 10)
	y := noShowPlayer(10, false, 1, now, 11)
	players = append(players, x, y)

	plan, err := matchmaking.PlanEvent(players, cfg, repairSettings())
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 2)
	roster, _ := planRosterAndSubIDs(plan)
	assert.True(t, roster[x.UserID])
	assert.True(t, roster[y.UserID])
}

func TestPlanEvent_ExcessiveCanSubStillBecomesSub(t *testing.T) {
	now := time.Now()
	cfg := matchmaking.Config{EventID: uuid.New(), TeamSize: 2, Slots: 4, TierCount: 25, SortLogic: "balanced", GameLabel: "Game 1 (2v2)", NoShowThreshold: 1}
	var players []matchmaking.Player
	for i := 0; i < 4; i++ {
		players = append(players, noShowPlayer(10, false, 0, now, i))
	}
	sub := noShowPlayer(10, true, 4, now, 10)
	players = append(players, sub)

	plan, err := matchmaking.PlanEvent(players, cfg, repairSettings())
	require.NoError(t, err)
	roster, subs := planRosterAndSubIDs(plan)
	assert.False(t, roster[sub.UserID])
	assert.True(t, subs[sub.UserID])
}

func TestPlanEvent_RankedModeExcludesExcessivePlayer(t *testing.T) {
	now := time.Now()
	cfg := matchmaking.Config{EventID: uuid.New(), TeamSize: 2, Slots: 4, TierCount: 25, SortLogic: "ranked", GameLabel: "Game 1 (2v2)", NoShowThreshold: 2}
	var players []matchmaking.Player
	for i := 0; i < 4; i++ {
		players = append(players, noShowPlayer(float64(10+i), false, 0, now, i))
	}
	// Mid-pack rank: ranked trimming would otherwise have kept them.
	excessive := noShowPlayer(11.5, false, 2, now, 10)
	players = append(players, excessive)

	plan, err := matchmaking.PlanEvent(players, cfg, repairSettings())
	require.NoError(t, err)
	roster, _ := planRosterAndSubIDs(plan)
	assert.False(t, roster[excessive.UserID])
	assert.Len(t, roster, 4)
}

func TestBestNonSubReplacement_SkipsDeprioritized(t *testing.T) {
	now := time.Now()
	vacated := player(uuid.New(), 10, true, now)
	exact := player(uuid.New(), 10, false, now.Add(time.Minute))
	exact.Deprioritized = true
	near := player(uuid.New(), 9, false, now.Add(2*time.Minute))

	got, ok := matchmaking.BestNonSubReplacementForTest([]matchmaking.Player{vacated, exact, near}, []matchmaking.LobbyPlan{{Roster: []matchmaking.Player{vacated}}}, vacated, 2, 25)
	require.True(t, ok)
	assert.Equal(t, near.UserID, got.UserID)
}

// unfairPlatBronzeLobby is the Asc/Asc/Plat/Bronze 2v2 the repair tests fix with an Iron leftover.
func unfairPlatBronzeLobby(t *testing.T, now time.Time) (matchmaking.LobbyPlan, []matchmaking.Player) {
	t.Helper()
	roster := []matchmaking.Player{
		player(uuid.New(), 21, false, now),
		player(uuid.New(), 21, false, now.Add(time.Minute)),
		player(uuid.New(), 13, false, now.Add(2*time.Minute)),
		player(uuid.New(), 6, false, now.Add(3*time.Minute)),
	}
	lobby := windowedLobby(t, roster, 2, 25)
	require.True(t, matchmaking.IsLobbyUnfair(lobby, repairSettings(), 25))
	return lobby, roster
}

func TestRepair_PrefersNonDeprioritizedLeftoverThatAlsoFixesLobby(t *testing.T) {
	now := time.Now()
	lobby, roster := unfairPlatBronzeLobby(t, now)
	// Registered first, so it would win every existing tie-break if absences were ignored.
	excessive := player(uuid.New(), 3, false, now.Add(-time.Hour))
	excessive.Deprioritized = true
	excessive.NoShowCount = 4
	clean := player(uuid.New(), 3, false, now.Add(10*time.Minute))

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		append(append([]matchmaking.Player{}, roster...), excessive, clean),
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[clean.UserID])
	assert.False(t, ids[excessive.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], repairSettings(), 25))
}

func TestRepair_UsesDeprioritizedLeftoverWhenNoOtherFixesLobby(t *testing.T) {
	now := time.Now()
	lobby, roster := unfairPlatBronzeLobby(t, now)
	excessive := player(uuid.New(), 3, false, now.Add(10*time.Minute))
	excessive.Deprioritized = true
	excessive.NoShowCount = 4
	// A Radiant leftover cannot help a lopsided 2v2.
	unhelpful := player(uuid.New(), 25, false, now.Add(11*time.Minute))

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		append(append([]matchmaking.Player{}, roster...), excessive, unhelpful),
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[excessive.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], repairSettings(), 25))
}

func TestRepair_DuoRequestDoesNotFavorDeprioritizedPartner(t *testing.T) {
	now := time.Now()
	lobby, roster := unfairPlatBronzeLobby(t, now)
	for i := range lobby.Roster {
		if lobby.Roster[i].AvgRank == 21 {
			lobby.Roster[i].DiscordName = "anchor"
			lobby.Roster[i].DuoRequest = duoRequest("partner")
			break
		}
	}
	partner := player(uuid.New(), 3, false, now.Add(-time.Hour))
	partner.DiscordName = "partner"
	partner.DuoRequest = duoRequest("anchor")
	partner.Deprioritized = true
	partner.NoShowCount = 5
	other := player(uuid.New(), 3, false, now.Add(10*time.Minute))

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		append(append([]matchmaking.Player{}, roster...), partner, other),
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[other.UserID])
	assert.False(t, ids[partner.UserID])
}

func TestPreferCleanRepair(t *testing.T) {
	// No clean trial: the overall best stands.
	assert.False(t, matchmaking.PreferCleanRepairForTest(false, 0, 0, false, false, 0, 0, true))
	// Clean trial leaves the lobby fair: it wins even when the overall best scores better.
	assert.True(t, matchmaking.PreferCleanRepairForTest(true, 2, 1, false, false, 0.5, 1, true))
	// Clean trial still unfair but ties the overall best on the optimized metric.
	assert.True(t, matchmaking.PreferCleanRepairForTest(true, 4, 1, true, false, 4, 0, true))
	// Clean trial still unfair and worse: a repeat no-show is the only way to a fairer match.
	assert.False(t, matchmaking.PreferCleanRepairForTest(true, 4, 1, true, false, 3.5, 1, true))
	// Outlier optimization compares the outlier gap instead.
	assert.True(t, matchmaking.PreferCleanRepairForTest(true, 9, 7, false, true, 1, 7, false))
}

func TestInPlaceTrialNotBetter_DeprioritizationBreaksFullTie(t *testing.T) {
	out := matchmaking.Player{UserID: uuid.New(), AvgRank: 13}
	clean := matchmaking.Player{UserID: uuid.New(), AvgRank: 3}
	fewer := matchmaking.Player{UserID: uuid.New(), AvgRank: 3, Deprioritized: true, NoShowCount: 3}
	more := matchmaking.Player{UserID: uuid.New(), AvgRank: 3, Deprioritized: true, NoShowCount: 6}

	notBetter := func(trialIn, bestIn matchmaking.Player) bool {
		return matchmaking.InPlaceTrialNotBetterForTest(true, 1, 1, 1, 1, 10, 10, out, trialIn, out, bestIn)
	}
	assert.True(t, notBetter(fewer, clean), "deprioritized does not beat a clean tie")
	assert.False(t, notBetter(clean, fewer), "clean beats a deprioritized tie")
	assert.False(t, notBetter(fewer, more), "fewer no-shows wins among deprioritized")
	assert.True(t, notBetter(more, fewer))
}

func TestRepair_PrefersFairCleanLeftoverOverBetterScoringDeprioritized(t *testing.T) {
	now := time.Now()
	lobby, roster := unfairPlatBronzeLobby(t, now)
	// Rank 6 would even the teams exactly; rank 10 only gets them within the fairness threshold.
	excessive := player(uuid.New(), 6, false, now.Add(10*time.Minute))
	excessive.Deprioritized = true
	excessive.NoShowCount = 4
	clean := player(uuid.New(), 10, false, now.Add(11*time.Minute))

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		append(append([]matchmaking.Player{}, roster...), excessive, clean),
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		repairSettings(),
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[clean.UserID])
	assert.False(t, ids[excessive.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], repairSettings(), 25))
}

func TestAddBackBefore_TieBreakOrder(t *testing.T) {
	now := time.Now()
	idLow := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	idHigh := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	base := matchmaking.Player{UserID: idLow, NoShowCount: 3, RegisteredGameCount: 1, CreatedAt: now}

	cases := []struct {
		name string
		a, b func(p matchmaking.Player) matchmaking.Player
		want bool
	}{
		{"fewer no-shows first", func(p matchmaking.Player) matchmaking.Player { p.NoShowCount = 2; return p }, nil, true},
		{"more no-shows later", func(p matchmaking.Player) matchmaking.Player { p.NoShowCount = 4; return p }, nil, false},
		{"fewer group games first", func(p matchmaking.Player) matchmaking.Player { p.RegisteredGameCount = 0; return p }, nil, true},
		{"more group games later", func(p matchmaking.Player) matchmaking.Player { p.RegisteredGameCount = 2; return p }, nil, false},
		{"earlier registration first", func(p matchmaking.Player) matchmaking.Player { p.CreatedAt = now.Add(-time.Minute); return p }, nil, true},
		{"later registration later", func(p matchmaking.Player) matchmaking.Player { p.CreatedAt = now.Add(time.Minute); return p }, nil, false},
		{"lower user ID first on full tie", nil, func(p matchmaking.Player) matchmaking.Player { p.UserID = idHigh; return p }, true},
		{"higher user ID later on full tie", func(p matchmaking.Player) matchmaking.Player { p.UserID = idHigh; return p }, nil, false},
		{"identical player is not before itself", nil, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, b := base, base
			if tc.a != nil {
				a = tc.a(a)
			}
			if tc.b != nil {
				b = tc.b(b)
			}
			assert.Equal(t, tc.want, matchmaking.AddBackBeforeForTest(a, b))
		})
	}
}

func TestAddsDeprioritized(t *testing.T) {
	clean := matchmaking.Player{UserID: uuid.New()}
	held := matchmaking.Player{UserID: uuid.New(), Deprioritized: true}

	assert.False(t, matchmaking.AddsDeprioritizedForTest(nil, nil), "empty combo adds nobody")
	assert.False(t, matchmaking.AddsDeprioritizedForTest([]matchmaking.Player{clean}, nil), "clean players only")
	assert.True(t, matchmaking.AddsDeprioritizedForTest([]matchmaking.Player{clean, held}, nil), "brings in a held-out leftover")
	assert.False(t, matchmaking.AddsDeprioritizedForTest([]matchmaking.Player{clean, held}, []uuid.UUID{held.UserID}),
		"a deprioritized player already in this window (e.g. added back) is not newly added")
}

func TestMarkDeprioritized_EnabledButNobodyAtThreshold(t *testing.T) {
	now := time.Now()
	players := []matchmaking.Player{noShowPlayer(10, false, 2, now, 0), noShowPlayer(10, false, 0, now, 1)}

	for _, p := range matchmaking.MarkDeprioritizedForTest(players, fiveVFiveCfg(3)) {
		assert.False(t, p.Deprioritized)
	}
}

func TestMarkDeprioritized_SecondPassSkipsPlayerAlreadyAddedBack(t *testing.T) {
	now := time.Now()
	cfg := matchmaking.Config{TeamSize: 2, Slots: 4, TierCount: 25, LobbyCount: 2, SubMin: 1, SortLogic: "balanced", NoShowThreshold: 3}
	var players []matchmaking.Player
	for i := 0; i < 4; i++ {
		players = append(players, noShowPlayer(10, false, 0, now, i))
	}
	for i := 0; i < 2; i++ {
		players = append(players, noShowPlayer(10, true, 0, now, 10+i))
	}
	nonSub := noShowPlayer(10, false, 3, now, 20) // fewest no-shows, restored by the non-sub pass
	canSubA := noShowPlayer(10, true, 4, now, 21)
	canSubB := noShowPlayer(10, true, 5, now, 22)
	players = append(players, nonSub, canSubA, canSubB)

	// 8 seats; 4 can-subs with 2 reserved → 6 non-subs needed. The non-sub pass restores the
	// only held non-sub (still short), then the total pass skips them and restores canSubA.
	flags := deprioritizedByID(matchmaking.MarkDeprioritizedForTest(players, cfg))
	assert.False(t, flags[nonSub.UserID])
	assert.False(t, flags[canSubA.UserID])
	assert.True(t, flags[canSubB.UserID])
}

// End-to-end outcome: the held-out player ends up seated whether the first pass picks them or a
// later pass swaps them in over the partial clean fix. The per-pass choice is pinned by TestPreferCleanRepair.
func TestRepair_UsesDeprioritizedLeftoverWhenCleanOnlyPartlyHelps(t *testing.T) {
	now := time.Now()
	// Tighter team-separation limit (1.0 at 25 tiers) so a partial improvement stays unfair.
	settings := matchmaking.Settings{FairnessOutlierGap: 6, FairnessTeamSeparation: 1, FairnessReferenceTierCount: 25}
	roster := []matchmaking.Player{
		player(uuid.New(), 21, false, now),
		player(uuid.New(), 21, false, now.Add(time.Minute)),
		player(uuid.New(), 13, false, now.Add(2*time.Minute)),
		player(uuid.New(), 6, false, now.Add(3*time.Minute)),
	}
	lobby := windowedLobby(t, roster, 2, 25)
	require.True(t, matchmaking.IsLobbyUnfair(lobby, settings, 25)) // teams 17 vs 13.5

	// Swapping rank 1 in for the Plat cuts separation from 3.5 to 2.5: accepted, but still unfair.
	clean := player(uuid.New(), 1, false, now.Add(-time.Hour))
	// Swapping rank 6 in for the Plat evens the teams exactly.
	excessive := player(uuid.New(), 6, false, now.Add(10*time.Minute))
	excessive.Deprioritized = true
	excessive.NoShowCount = 4

	out := matchmaking.RepairUnfairWindowPairsForTest(
		[]matchmaking.LobbyPlan{lobby},
		append(append([]matchmaking.Player{}, roster...), clean, excessive),
		matchmaking.Config{TeamSize: 2, TierCount: 25, LobbyCount: 1, SortLogic: "balanced"},
		settings,
	)
	ids := rosterIDSet(out[0])
	assert.True(t, ids[excessive.UserID], "the held-out player is the only way to a fair match")
	assert.False(t, ids[clean.UserID])
	assert.False(t, matchmaking.IsLobbyUnfair(out[0], settings, 25))
}
