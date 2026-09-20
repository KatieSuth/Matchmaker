package main

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
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
