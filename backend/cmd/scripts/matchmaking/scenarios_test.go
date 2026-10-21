package main

import (
	"math/bits"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatScenarioGroupName(t *testing.T) {
	t.Parallel()

	t.Run("joins sort logic and use case", func(t *testing.T) {
		got, err := formatScenarioGroupName("balanced", "9/10 insufficient players")
		require.NoError(t, err)
		assert.Equal(t, "balanced: 9/10 insufficient players", got)
	})

	t.Run("allows exactly 50 runes", func(t *testing.T) {
		useCase := strings.Repeat("a", eventGroupNameMaxRunes-len("ranked: "))
		got, err := formatScenarioGroupName("ranked", useCase)
		require.NoError(t, err)
		assert.Equal(t, eventGroupNameMaxRunes, utf8.RuneCountInString(got))
	})

	t.Run("rejects over 50 runes", func(t *testing.T) {
		useCase := strings.Repeat("a", eventGroupNameMaxRunes-len("balanced: ")+1)
		_, err := formatScenarioGroupName("balanced", useCase)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "max 50")
	})
}

func TestScenarioTemplatesGroupNamesFitLimit(t *testing.T) {
	t.Parallel()

	scenarios := allScenarios(stubSystemGames())
	require.NotEmpty(t, scenarios)

	seen := map[string]string{}
	for _, sc := range scenarios {
		require.NotEmpty(t, sc.Name, "scenario %s missing use-case name", sc.Key)
		name, err := formatScenarioGroupName(sc.SortLogic, sc.Name)
		require.NoError(t, err, "scenario %s/%s", sc.Key, sc.SortLogic)
		assert.LessOrEqual(t, utf8.RuneCountInString(name), eventGroupNameMaxRunes)
		assert.True(t, strings.HasPrefix(name, sc.SortLogic+": "))

		if other := seen[name]; other != "" {
			t.Errorf("duplicate event group name %q for %s and %s", name, other, sc.Key+"/"+sc.SortLogic)
		}
		seen[name] = sc.Key + "/" + sc.SortLogic
	}
}

func TestValorantLadderRegs(t *testing.T) {
	t.Parallel()

	regs := valorantLadderRegs()
	require.Len(t, regs, valorantLadderSize)

	var ironPlat, diamond, ascendant, immortal, radiant, canSub int
	slots := map[int]bool{}
	for _, r := range regs {
		assert.False(t, slots[r.Slot], "duplicate slot %d", r.Slot)
		slots[r.Slot] = true
		switch {
		case r.RankOrder >= valRankIron1 && r.RankOrder <= valRankPlat3:
			ironPlat++
		case r.RankOrder >= valRankDiamond1 && r.RankOrder <= valRankDiamond3:
			diamond++
		case r.RankOrder >= valRankAscendant1 && r.RankOrder <= valRankAscendant3:
			ascendant++
		case r.RankOrder >= valRankImmortal1 && r.RankOrder < valRankRadiant:
			immortal++
		case r.RankOrder == valRankRadiant:
			radiant++
		default:
			t.Errorf("slot %d has unexpected rank order %d", r.Slot, r.RankOrder)
		}
		if r.CanSub {
			canSub++
		}
	}

	assert.Equal(t, 18, ironPlat, "60%% Iron–Platinum")
	assert.Equal(t, 6, diamond, "20%% Diamond")
	assert.Equal(t, 3, ascendant, "10%% Ascendant")
	assert.Equal(t, 2, immortal, "~5%% Immortal (rounded)")
	assert.Equal(t, 1, radiant, "~5%% Radiant (rounded)")
	assert.Equal(t, valorantLadderSubCount, canSub)
}

func TestSharedLadderEvents(t *testing.T) {
	t.Parallel()

	events := sharedLadderEvents(modeName5v5, modeName3v3Skirmish)
	require.Len(t, events, 2)
	assert.Equal(t, modeName5v5, events[0].ModeName)
	assert.Equal(t, modeName3v3Skirmish, events[1].ModeName)
	require.Len(t, events[0].Registrations, valorantLadderSize)
	require.Len(t, events[1].Registrations, valorantLadderSize)

	assert.Equal(t, events[0].Registrations, events[1].Registrations, "both rows should register the same users")
	events[0].Registrations[0].CanSub = !events[0].Registrations[0].CanSub
	assert.NotEqual(t, events[0].Registrations[0].CanSub, events[1].Registrations[0].CanSub, "event rows must not share a backing slice")
}

func TestMultiGameLadderScenarios(t *testing.T) {
	t.Parallel()

	byKey := map[string]Scenario{}
	for _, sc := range scenarioTemplates(stubSystemGames()) {
		byKey[sc.Key] = sc
	}

	same, ok := byKey["multi_same_mode_ladder"]
	require.True(t, ok)
	assert.Equal(t, int32(3), same.SubMin)
	assert.Equal(t, gameKeyValorant, same.GameKey)
	require.Len(t, same.Events, 2)
	assert.Equal(t, modeName5v5, same.Events[0].ModeName)
	assert.Equal(t, modeName5v5, same.Events[1].ModeName)
	assert.Len(t, same.Events[0].Registrations, valorantLadderSize)
	assert.Len(t, same.Events[1].Registrations, valorantLadderSize)
	assert.Equal(t, same.Events[0].Registrations, same.Events[1].Registrations)

	diff, ok := byKey["multi_diff_mode_ladder"]
	require.True(t, ok)
	assert.Equal(t, int32(3), diff.SubMin)
	assert.Equal(t, gameKeyValorant, diff.GameKey)
	require.Len(t, diff.Events, 2)
	assert.Equal(t, modeName5v5, diff.Events[0].ModeName)
	assert.Equal(t, modeName3v3Skirmish, diff.Events[1].ModeName)
	assert.Len(t, diff.Events[0].Registrations, valorantLadderSize)
	assert.Len(t, diff.Events[1].Registrations, valorantLadderSize)
	assert.Equal(t, diff.Events[0].Registrations, diff.Events[1].Registrations)
}

// stubSystemGames supplies rank orders so scenarioTemplates can build without a database.
func stubSystemGames() *systemGames {
	ranks := []db.GameRank{{Order: 27}}
	return &systemGames{
		Valorant: gameInfo{Ranks: ranks},
		LoL:      gameInfo{Ranks: ranks},
	}
}

// planNoShowScenario runs a scenario's single event through matchmaking the way lock-in would:
// Valorant's 25 tiers, .env.example fairness defaults, and registration order by slot.
func planNoShowScenario(t *testing.T, sc Scenario, sortLogic string) (rostered map[int]bool, unfair bool) {
	t.Helper()
	require.Len(t, sc.Events, 1)
	base := time.Now()
	slotByID := map[uuid.UUID]int{}
	var players []matchmaking.Player
	for _, reg := range sc.Events[0].Registrations {
		id := uuid.New()
		slotByID[id] = reg.Slot
		players = append(players, matchmaking.Player{
			UserID:        id,
			AvgRank:       float64(reg.RankOrder),
			CanSubstitute: reg.CanSub,
			CanLobbyHost:  reg.CanHost,
			CreatedAt:     base.Add(time.Duration(reg.Slot) * time.Second),
			NoShowCount:   reg.NoShows,
		})
	}
	cfg := matchmaking.Config{
		EventID: uuid.New(), TeamSize: 5, Slots: 10, TierCount: 25, SubMin: int(sc.SubMin),
		SortLogic: sortLogic, GameLabel: sc.Key, NoShowThreshold: int(sc.NoShowThreshold),
	}
	settings := matchmaking.Settings{FairnessOutlierGap: 6, FairnessTeamSeparation: 3, FairnessReferenceTierCount: 25}
	plan, err := matchmaking.PlanEvent(players, cfg, settings)
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)

	rostered = map[int]bool{}
	for _, p := range plan.Lobbies[0].Roster {
		rostered[slotByID[p.UserID]] = true
	}
	return rostered, plan.Lobbies[0].FairnessWarning
}

func TestNoShowScenarios_ProduceIntendedLockIn(t *testing.T) {
	t.Parallel()

	byKey := map[string]Scenario{}
	for _, sc := range scenarioTemplates(stubSystemGames()) {
		byKey[sc.Key] = sc
	}

	for _, sortLogic := range []string{"balanced", "ranked"} {
		t.Run(sortLogic, func(t *testing.T) {
			excluded := byKey["noshow_excluded"]
			assert.Equal(t, int32(noShowScenarioThreshold), excluded.NoShowThreshold)
			rostered, unfair := planNoShowScenario(t, excluded, sortLogic)
			assert.False(t, rostered[0], "3 no-shows meets the limit")
			assert.False(t, rostered[1], "5 no-shows exceeds the limit")
			assert.Len(t, rostered, 10)
			assert.False(t, unfair)

			addBack := byKey["noshow_add_back"]
			rostered, unfair = planNoShowScenario(t, addBack, sortLogic)
			assert.False(t, rostered[0], "4 no-shows stays out")
			assert.True(t, rostered[1], "3 no-shows is added back to reach 10")
			assert.False(t, unfair)

			fairness := byKey["noshow_fairness"]
			rostered, unfair = planNoShowScenario(t, fairness, sortLogic)
			assert.True(t, rostered[0], "only the held-out player can make the lobby fair")
			assert.False(t, unfair)

			// Prove "must" independently of the algorithm: no 10 of the clean players, split any
			// way, passes both fairness checks.
			assert.False(t, anyFairLobby(fairness.Events[0].Registrations[1:]), "a fair lobby exists without the held-out player")
		})
	}
}

func TestMaxScenarioNoShows(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, maxScenarioNoShows(Scenario{Events: []EventSpec{{Registrations: flatRankRegs(3, 10)}}}))
	assert.Equal(t, 5, maxScenarioNoShows(Scenario{Events: []EventSpec{
		{Registrations: withNoShows(flatRankRegs(3, 10), map[int]int{0: 2})},
		{Registrations: withNoShows(flatRankRegs(3, 10), map[int]int{1: 5})},
	}}))
	assert.Equal(t, 0, maxScenarioNoShows(Scenario{}), "no events")
}

func TestNoShowHistoryGroupNamesFitLimit(t *testing.T) {
	t.Parallel()

	for _, sc := range allScenarios(stubSystemGames()) {
		if maxScenarioNoShows(sc) == 0 {
			continue
		}
		ids := map[uuid.UUID]bool{scenarioGroupID(sc.Key, sc.SortLogic): true}
		for i := 0; i < maxScenarioNoShows(sc); i++ {
			_, err := formatScenarioGroupName(sc.SortLogic, noShowHistoryGroupUseCase(sc, i))
			require.NoError(t, err, sc.Key)
			id := noShowHistoryGroupID(sc, i)
			assert.False(t, ids[id], "%s history group %d needs its own ID", sc.Key, i)
			ids[id] = true
		}
	}
}

func TestNoShowHistoryGroupID_IndexZeroKeepsLegacyID(t *testing.T) {
	t.Parallel()

	sc := Scenario{Key: "noshow_excluded", SortLogic: "balanced"}
	assert.Equal(t, scenarioGroupID("noshow_excluded_history", "balanced"), noShowHistoryGroupID(sc, 0),
		"cleanup must still find groups seeded with the single-history-group layout")
	assert.NotEqual(t, noShowHistoryGroupID(sc, 0), noShowHistoryGroupID(sc, 1))
}

// anyFairLobby brute-forces every choice of 10 registrants and every 5/5 split, reporting whether
// any passes both fairness checks at Valorant's 25 tiers with the .env.example defaults.
func anyFairLobby(regs []RegSpec) bool {
	settings := matchmaking.Settings{FairnessOutlierGap: 6, FairnessTeamSeparation: 3, FairnessReferenceTierCount: 25}
	n := len(regs)
	for choose := 0; choose < 1<<n; choose++ {
		if bits.OnesCount(uint(choose)) != 10 {
			continue
		}
		var ten []RegSpec
		for i := range regs {
			if choose&(1<<i) != 0 {
				ten = append(ten, regs[i])
			}
		}
		// Fix ten[0] on team 1 so each split is tried once.
		for split := 1; split < 1<<10; split += 2 {
			if bits.OnesCount(uint(split)) != 5 {
				continue
			}
			roster := make([]matchmaking.Player, 0, 10)
			for i, reg := range ten {
				team := 2
				if split&(1<<i) != 0 {
					team = 1
				}
				roster = append(roster, matchmaking.Player{UserID: uuid.New(), AvgRank: float64(reg.RankOrder), TeamNumber: &team})
			}
			if !matchmaking.IsLobbyUnfair(matchmaking.LobbyPlan{Roster: roster}, settings, 25) {
				return true
			}
		}
	}
	return false
}

func TestAnyFairLobby(t *testing.T) {
	t.Parallel()
	assert.True(t, anyFairLobby(flatRankRegs(11, 10)), "equal ranks are always fair")
	assert.False(t, anyFairLobby(regsFromSlotRanks(
		[2]int{0, 25}, [2]int{1, 1}, [2]int{2, 1}, [2]int{3, 1}, [2]int{4, 1},
		[2]int{5, 1}, [2]int{6, 1}, [2]int{7, 1}, [2]int{8, 1}, [2]int{9, 1},
	)), "a lone Radiant among Irons can never be fair")
}

// planScenarioLobbies plans a scenario's first event the way lock-in would (duo requests
// included) and returns the lobby index and team of each slot.
func planScenarioLobbies(t *testing.T, sc Scenario, sortLogic string, teamSize int) (lobbyOf, teamOf map[int]int) {
	t.Helper()
	base := time.Now()
	slotByID := map[uuid.UUID]int{}
	var players []matchmaking.Player
	for _, reg := range sc.Events[0].Registrations {
		id := uuid.New()
		slotByID[id] = reg.Slot
		var duo *string
		if reg.DuoPartner != nil {
			name := scenarioDiscordName(sc.Key, sortLogic, *reg.DuoPartner)
			duo = &name
		}
		players = append(players, matchmaking.Player{
			UserID:        id,
			DiscordName:   scenarioDiscordName(sc.Key, sortLogic, reg.Slot),
			DuoRequest:    duo,
			AvgRank:       float64(reg.RankOrder),
			CanSubstitute: reg.CanSub,
			CreatedAt:     base.Add(time.Duration(reg.Slot) * time.Second),
		})
	}
	plan, err := matchmaking.PlanEvent(players, matchmaking.Config{
		EventID: uuid.New(), TeamSize: teamSize, Slots: 2 * teamSize, TierCount: 25, SubMin: int(sc.SubMin),
		SortLogic: sortLogic, GameLabel: sc.Key,
	}, matchmaking.Settings{FairnessOutlierGap: 6, FairnessTeamSeparation: 3, FairnessReferenceTierCount: 25})
	require.NoError(t, err)
	lobbyOf, teamOf = map[int]int{}, map[int]int{}
	for i, lobby := range plan.Lobbies {
		for _, p := range lobby.Roster {
			lobbyOf[slotByID[p.UserID]] = i
			teamOf[slotByID[p.UserID]] = *p.TeamNumber
		}
	}
	return lobbyOf, teamOf
}

func TestDuoScenarios_MatchTheirDescriptions(t *testing.T) {
	t.Parallel()
	byKey := map[string]Scenario{}
	for _, sc := range scenarioTemplates(stubSystemGames()) {
		byKey[sc.Key] = sc
	}

	cases := []struct {
		key                 string
		a, b                int
		sameLobby, sameTeam bool
	}{
		{"duos_lobby_unite", 4, 14, true, true},
		{"duos_lobby_blocked", 0, 1, false, false},
		{"duos_team_unite", 1, 3, true, true},
		{"duos_team_blocked", 0, 2, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			sc, ok := byKey[tc.key]
			require.True(t, ok)
			lobbyOf, teamOf := planScenarioLobbies(t, sc, "balanced", 5)
			require.Contains(t, lobbyOf, tc.a)
			require.Contains(t, lobbyOf, tc.b)
			sameLobby := lobbyOf[tc.a] == lobbyOf[tc.b]
			assert.Equal(t, tc.sameLobby, sameLobby, "same lobby")
			assert.Equal(t, tc.sameTeam, sameLobby && teamOf[tc.a] == teamOf[tc.b], "same team")
		})
	}
}

func TestNoShowViewerScenarios(t *testing.T) {
	t.Parallel()
	byKey := map[string]Scenario{}
	for _, sc := range scenarioTemplates(stubSystemGames()) {
		byKey[sc.Key] = sc
	}

	cases := []struct {
		key       string
		excessive bool // red warning (at/above the limit) vs yellow notice
	}{
		{"noshow_viewer", true},
		{"noshow_viewer_under", false},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			sc, ok := byKey[tc.key]
			require.True(t, ok)
			assert.Positive(t, sc.ViewerNoShows, "the --host user has no-shows")
			assert.Positive(t, sc.NoShowThreshold, "deprioritization is on, so the banner shows")
			assert.Equal(t, tc.excessive, sc.ViewerNoShows >= int(sc.NoShowThreshold))
			assert.Equal(t, sc.ViewerNoShows, maxScenarioNoShows(sc), "history holds enough past games for the viewer's marks")
		})
	}
	assert.GreaterOrEqual(t, byKey["noshow_viewer"].ViewerNoShows, 3, "at least 3 no-shows, as requested")
}

func TestScenarioPlayerSlotsStayBelowSeededHost(t *testing.T) {
	t.Parallel()
	for _, sc := range allScenarios(stubSystemGames()) {
		for _, ev := range sc.Events {
			for _, reg := range ev.Registrations {
				assert.Less(t, reg.Slot, scenarioHostSlot, "%s slot %d would collide with the seeded host", sc.Key, reg.Slot)
			}
		}
	}
}
