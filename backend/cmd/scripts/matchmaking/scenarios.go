package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/cmd/scripts/common"
	"github.com/google/uuid"
)

type RegSpec struct {
	Slot       int
	RankOrder  int
	CanSub     bool
	CanHost    bool
	DuoPartner *int
}

type EventSpec struct {
	ModeName      string // e.g. "5v5", "3v3 Skirmish" — must belong to the scenario's game
	Registrations []RegSpec
}

type Scenario struct {
	Key         string
	SortLogic   string
	GameKey     string // single game for the entire group: "valorant" | "lol"
	SubMin      int32
	Region      string
	Description string
	Events      []EventSpec
}

type SeedResult struct {
	Key         string `json:"key"`
	SortLogic   string `json:"sort_logic"`
	Game        string `json:"game"`
	GroupID     string `json:"group_id"`
	Description string `json:"description"`
	Skipped     bool   `json:"skipped"`
}

const sharedMultiEventUsers = 10

func spreadRankRegs(count, rankStart int, opts func(slot int) RegSpec) []RegSpec {
	out := make([]RegSpec, count)
	for i := 0; i < count; i++ {
		spec := opts(i)
		spec.Slot = i
		if spec.RankOrder == 0 {
			spec.RankOrder = rankStart + i
		}
		out[i] = spec
	}
	return out
}

func withSubSlots(regs []RegSpec, subSlots map[int]bool) []RegSpec {
	out := make([]RegSpec, len(regs))
	copy(out, regs)
	for i := range out {
		if subSlots[i] {
			out[i].CanSub = true
		}
	}
	return out
}

func subSlotSet(count int, subIndices ...int) map[int]bool {
	m := make(map[int]bool, len(subIndices))
	for _, i := range subIndices {
		if i >= 0 && i < count {
			m[i] = true
		}
	}
	return m
}

// linkDuoPairsBySlot sets mutual duo_request targets using player slot numbers.
func linkDuoPairsBySlot(regs []RegSpec, pairs ...[2]int) []RegSpec {
	slotToIdx := make(map[int]int, len(regs))
	for i, r := range regs {
		slotToIdx[r.Slot] = i
	}
	out := make([]RegSpec, len(regs))
	copy(out, regs)
	for _, p := range pairs {
		slotA, slotB := p[0], p[1]
		idxA, okA := slotToIdx[slotA]
		idxB, okB := slotToIdx[slotB]
		if !okA || !okB {
			common.Fatal("duo pair references unknown slot", "slot_a", slotA, "slot_b", slotB)
		}
		out[idxA].DuoPartner = &slotB
		out[idxB].DuoPartner = &slotA
	}
	return out
}

// regsFromSlotRanks builds registrations with explicit slot/rank pairs.
func regsFromSlotRanks(pairs ...[2]int) []RegSpec {
	regs := make([]RegSpec, len(pairs))
	for i, p := range pairs {
		regs[i] = RegSpec{Slot: p[0], RankOrder: p[1]}
	}
	return regs
}

// oneSidedDuo sets a duo request on fromSlot targeting toSlot without a mutual request back.
func oneSidedDuo(regs []RegSpec, fromSlot, toSlot int) []RegSpec {
	out := make([]RegSpec, len(regs))
	copy(out, regs)
	for i := range out {
		if out[i].Slot == fromSlot {
			out[i].DuoPartner = &toSlot
			break
		}
	}
	return out
}

// multiEventRegistrations builds registrations for a multi-event group. The first
// sharedMultiEventUsers slots (0–9) are the same users across every event row;
// additional registrants use event-specific slots so each row does not get its own
// full duplicate user pool.
func multiEventRegistrations(eventIndex, totalCount, rankStart int, opts func(slot int) RegSpec) []RegSpec {
	if totalCount <= sharedMultiEventUsers {
		return spreadRankRegs(totalCount, rankStart, opts)
	}

	regs := spreadRankRegs(sharedMultiEventUsers, rankStart, func(slot int) RegSpec {
		return RegSpec{}
	})
	extraCount := totalCount - sharedMultiEventUsers
	extraSlotStart := sharedMultiEventUsers + (eventIndex * extraCount)
	for i := 0; i < extraCount; i++ {
		slot := extraSlotStart + i
		spec := opts(slot)
		spec.Slot = slot
		if spec.RankOrder == 0 {
			spec.RankOrder = rankStart + sharedMultiEventUsers + i
		}
		regs = append(regs, spec)
	}
	return regs
}

// uniqueMultiEventRegistrations assigns slots starting at slotOffset so each event
// row gets a completely separate user pool (no shared registrants).
func uniqueMultiEventRegistrations(slotOffset, count, rankStart int, opts func(slot int) RegSpec) []RegSpec {
	out := make([]RegSpec, count)
	for i := 0; i < count; i++ {
		slot := slotOffset + i
		spec := opts(slot)
		spec.Slot = slot
		if spec.RankOrder == 0 {
			spec.RankOrder = rankStart + i
		}
		out[i] = spec
	}
	return out
}

func scenarioTemplates(g *systemGames) []Scenario {
	maxValRank := int(g.Valorant.Ranks[len(g.Valorant.Ranks)-1].Order)
	maxLoLRank := int(g.LoL.Ranks[len(g.LoL.Ranks)-1].Order)

	build := func(key, gameKey, region, description string, subMin int32, events []EventSpec) Scenario {
		return Scenario{
			Key:         key,
			GameKey:     gameKey,
			SubMin:      subMin,
			Region:      region,
			Description: description,
			Events:      events,
		}
	}

	valEvent := func(regs []RegSpec) []EventSpec {
		return []EventSpec{{ModeName: modeName5v5, Registrations: regs}}
	}
	lolEvent := func(regs []RegSpec) []EventSpec {
		return []EventSpec{{ModeName: modeName5v5, Registrations: regs}}
	}
	valEvents := func(events ...EventSpec) []EventSpec {
		out := make([]EventSpec, len(events))
		copy(out, events)
		for i := range out {
			out[i].ModeName = defaultModeName(out[i].ModeName)
		}
		return out
	}

	templates := []Scenario{
		build("insufficient_players", gameKeyValorant, "AMER",
			"9/10 players — lock-in should fail with insufficient players",
			0, valEvent(spreadRankRegs(9, 12, func(slot int) RegSpec {
				return RegSpec{RankOrder: 12}
			}))),
		build("single_lobby", gameKeyValorant, "AMER",
			"Exactly 10 players — expect 1 lobby at 5v5 capacity",
			0, valEvent(spreadRankRegs(10, 8, func(slot int) RegSpec {
				return RegSpec{}
			}))),
		build("single_lobby_overflow", gameKeyValorant, "EMEA",
			"14 players — 1 lobby, 2 overflow subs, 2 unplaced",
			0, valEvent(spreadRankRegs(14, 8, func(slot int) RegSpec {
				switch {
				case slot >= 12:
					return RegSpec{CanSub: false}
				case slot >= 10:
					return RegSpec{CanSub: true}
				default:
					return RegSpec{}
				}
			}))),
		build("two_lobbies_no_subs", gameKeyValorant, "APAC",
			"20 players, sub_min=0 — expect 2 lobbies",
			0, valEvent(spreadRankRegs(20, 6, func(slot int) RegSpec {
				return RegSpec{}
			}))),
		build("two_lobbies_mandatory_subs", gameKeyValorant, "AMER",
			"26 players with 6 subs, sub_min=3 — expect 2 lobbies with mandatory subs",
			3, valEvent(withSubSlots(spreadRankRegs(26, 5, func(slot int) RegSpec {
				return RegSpec{}
			}), subSlotSet(26, 20, 21, 22, 23, 24, 25)))),
		build("insufficient_substitutes", gameKeyValorant, "EMEA",
			"26 players but only 5 subs with sub_min=3 — capped at 1 lobby",
			3, valEvent(withSubSlots(spreadRankRegs(26, 5, func(slot int) RegSpec {
				return RegSpec{}
			}), subSlotSet(26, 0, 1, 2, 3, 4)))),
		build("roster_trim", gameKeyValorant, "APAC",
			"15 players — 10 rostered, 5 unplaced after lock-in",
			0, valEvent(spreadRankRegs(15, 1, func(slot int) RegSpec {
				return RegSpec{}
			}))),
		build("sub_capacity_swaps", gameKeyValorant, "AMER",
			"28 players (~18 subs), sub_min=3 — sub-cap roster swaps across 2 lobbies",
			3, valEvent(withSubSlots(spreadRankRegs(28, 4, func(slot int) RegSpec {
				return RegSpec{}
			}), subSlotSet(28, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17)))),
		build("overflow_subs", gameKeyValorant, "EMEA",
			"24 players (8 subs), sub_min=2 — 2 lobbies with overflow subs",
			2, valEvent(withSubSlots(spreadRankRegs(24, 6, func(slot int) RegSpec {
				return RegSpec{}
			}), subSlotSet(24, 16, 17, 18, 19, 20, 21, 22, 23)))),
		build("outlier_fairness", gameKeyValorant, "APAC",
			"9 mid ranks + 1 high outlier — expect fairness warning (outlier gap)",
			0, valEvent(func() []RegSpec {
				regs := spreadRankRegs(10, 12, func(slot int) RegSpec {
					return RegSpec{RankOrder: 12}
				})
				regs[0].RankOrder = maxValRank
				return regs
			}())),
		build("team_sep_fairness", gameKeyValorant, "AMER",
			"5 high + 5 low ranks — expect fairness warning (team separation)",
			0, valEvent(spreadRankRegs(10, 0, func(slot int) RegSpec {
				if slot < 5 {
					return RegSpec{RankOrder: 22}
				}
				return RegSpec{RankOrder: 3}
			}))),
		build("mutual_duos", gameKeyValorant, "EMEA",
			"22 players, 2 lobbies, 2 mutual duos — balance may block uniting both pairs",
			0, valEvent(func() []RegSpec {
				regs := spreadRankRegs(22, 6, func(slot int) RegSpec {
					return RegSpec{}
				})
				return linkDuoPairsBySlot(regs, [2]int{0, 1}, [2]int{10, 11})
			}())),
		build("duos_lobby_unite", gameKeyValorant, "AMER",
			"20 players, 2 lobbies — 1 close-rank duo should unite when spread allows",
			0, valEvent(linkDuoPairsBySlot(
				spreadRankRegs(20, 8, func(slot int) RegSpec { return RegSpec{} }),
				[2]int{4, 14},
			))),
		build("duos_lobby_blocked", gameKeyValorant, "EMEA",
			"20 players, 2 lobbies — high/low split duo should stay apart to preserve balance",
			0, valEvent(linkDuoPairsBySlot(regsFromSlotRanks(
				[2]int{0, 24}, [2]int{1, 6}, [2]int{2, 23}, [2]int{3, 6}, [2]int{4, 22}, [2]int{5, 6},
				[2]int{6, 21}, [2]int{7, 6}, [2]int{8, 20}, [2]int{9, 6},
				[2]int{10, 22}, [2]int{11, 4}, [2]int{12, 21}, [2]int{13, 4}, [2]int{14, 20}, [2]int{15, 4},
				[2]int{16, 19}, [2]int{17, 4}, [2]int{18, 18}, [2]int{19, 4},
			), [2]int{0, 10}))),
		build("duos_three_lobbies", gameKeyValorant, "APAC",
			"30 players, 3 lobbies — 3 mutual duos with close ranks each",
			0, valEvent(linkDuoPairsBySlot(
				spreadRankRegs(30, 6, func(slot int) RegSpec { return RegSpec{} }),
				[2]int{2, 3}, [2]int{12, 13}, [2]int{22, 23},
			))),
		build("duos_four_lobbies", gameKeyValorant, "AMER",
			"40 players, 4 lobbies — 4 mutual duos across the pool",
			0, valEvent(linkDuoPairsBySlot(
				spreadRankRegs(40, 5, func(slot int) RegSpec { return RegSpec{} }),
				[2]int{0, 1}, [2]int{10, 11}, [2]int{20, 21}, [2]int{30, 31},
			))),
		build("duos_team_unite", gameKeyValorant, "EMEA",
			"10 players, 1 lobby — close-rank duo should land on same team without worsening balance",
			0, valEvent(linkDuoPairsBySlot(
				spreadRankRegs(10, 8, func(slot int) RegSpec { return RegSpec{} }),
				[2]int{1, 3},
			))),
		build("duos_team_blocked", gameKeyValorant, "APAC",
			"10 players, 1 lobby — high-skill duo kept apart on teams to preserve balance",
			0, valEvent(linkDuoPairsBySlot(regsFromSlotRanks(
				[2]int{0, 22}, [2]int{1, 10}, [2]int{2, 20}, [2]int{3, 18},
				[2]int{4, 16}, [2]int{5, 14}, [2]int{6, 12}, [2]int{7, 10},
				[2]int{8, 8}, [2]int{9, 6},
			), [2]int{0, 2}))),
		build("duos_one_sided", gameKeyValorant, "AMER",
			"20 players, 2 lobbies — one-sided duo request ignored (no mutual match)",
			0, valEvent(oneSidedDuo(
				spreadRankRegs(20, 6, func(slot int) RegSpec { return RegSpec{} }),
				0, 10,
			))),
		build("duos_max_one_per_team", gameKeyValorant, "EMEA",
			"10 players, 1 lobby — 2 duos requested; at most 1 duo honored per team",
			0, valEvent(linkDuoPairsBySlot(regsFromSlotRanks(
				[2]int{0, 20}, [2]int{1, 19}, [2]int{2, 18}, [2]int{3, 17},
				[2]int{4, 12}, [2]int{5, 11}, [2]int{6, 10}, [2]int{7, 9},
				[2]int{8, 8}, [2]int{9, 7},
			), [2]int{0, 1}, [2]int{2, 3}))),
		build("lobby_host", gameKeyValorant, "APAC",
			"can_lobby_host on slot 3 — host selection priority",
			0, valEvent(spreadRankRegs(10, 10, func(slot int) RegSpec {
				return RegSpec{CanHost: slot == 3}
			}))),
		build("lol_outlier_fairness", gameKeyLoL, "AMER",
			"LoL tier scaling — 9 mid ranks + 1 high outlier fairness warning",
			0, lolEvent(func() []RegSpec {
				regs := spreadRankRegs(10, 12, func(slot int) RegSpec {
					return RegSpec{RankOrder: 12}
				})
				regs[0].RankOrder = maxLoLRank
				return regs
			}())),
		build("multi_game_skip_empty", gameKeyValorant, "EMEA",
			"Valorant: 5v5 with 10 players + 3v3 Skirmish row with 0 registrations — second event skipped",
			0, valEvents(
				EventSpec{ModeName: modeName5v5, Registrations: spreadRankRegs(10, 8, func(slot int) RegSpec {
					return RegSpec{}
				})},
				EventSpec{ModeName: modeName3v3Skirmish, Registrations: nil},
			)),
		build("multi_game_mixed_counts", gameKeyValorant, "APAC",
			"Valorant: 5v5 with 10 players + second 5v5 row with 20 (10 shared users)",
			0, valEvents(
				EventSpec{ModeName: modeName5v5, Registrations: multiEventRegistrations(0, 10, 8, func(slot int) RegSpec {
					return RegSpec{}
				})},
				EventSpec{ModeName: modeName5v5, Registrations: multiEventRegistrations(1, 20, 6, func(slot int) RegSpec {
					return RegSpec{}
				})},
			)),
		build("multi_game_insufficient_one", gameKeyValorant, "AMER",
			"Valorant: 5v5 with 10 players (1 lobby) + second 5v5 row with 9 — lock-in fails",
			0, valEvents(
				EventSpec{ModeName: modeName5v5, Registrations: multiEventRegistrations(0, 10, 8, func(slot int) RegSpec {
					return RegSpec{}
				})},
				EventSpec{ModeName: modeName5v5, Registrations: multiEventRegistrations(1, 9, 6, func(slot int) RegSpec {
					return RegSpec{}
				})},
			)),
		build("multi_game_two_lobbies_subs", gameKeyValorant, "EMEA",
			"Valorant: two 5v5 rows each with 30 players (10 shared), sub_min=5 — 2 lobbies per event",
			5, valEvents(
				EventSpec{ModeName: modeName5v5, Registrations: withSubSlots(
					multiEventRegistrations(0, 30, 5, func(slot int) RegSpec { return RegSpec{} }),
					subSlotSet(50, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29),
				)},
				EventSpec{ModeName: modeName5v5, Registrations: withSubSlots(
					multiEventRegistrations(1, 30, 5, func(slot int) RegSpec { return RegSpec{} }),
					subSlotSet(50, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49),
				)},
			)),
		build("multi_game_unique_users", gameKeyValorant, "APAC",
			"Valorant: two 5v5 rows with 10 players each — disjoint user pools (no overlap)",
			0, valEvents(
				EventSpec{ModeName: modeName5v5, Registrations: uniqueMultiEventRegistrations(0, 10, 8, func(slot int) RegSpec {
					return RegSpec{}
				})},
				EventSpec{ModeName: modeName5v5, Registrations: uniqueMultiEventRegistrations(10, 10, 8, func(slot int) RegSpec {
					return RegSpec{}
				})},
			)),
	}

	return templates
}

func allScenarios(g *systemGames) []Scenario {
	templates := scenarioTemplates(g)
	var out []Scenario
	for _, sortLogic := range []string{"balanced", "ranked"} {
		for _, tmpl := range templates {
			sc := tmpl
			sc.SortLogic = sortLogic
			sc.Key = tmpl.Key
			out = append(out, sc)
		}
	}
	return out
}

func registerEventPlayers(
	seed *common.SeedContext,
	sc Scenario,
	eventID uuid.UUID,
	game *gameInfo,
	specs []RegSpec,
) map[int]string {
	names := make(map[int]string, len(specs))
	for _, spec := range specs {
		_, name := createScenarioPlayer(seed, sc.Key, sc.SortLogic, spec.Slot, game, spec.RankOrder)
		names[spec.Slot] = name
	}

	for _, spec := range specs {
		userID := scenarioUserID(sc.Key, sc.SortLogic, spec.Slot)
		var duoRequest *string
		if spec.DuoPartner != nil {
			partnerName := names[*spec.DuoPartner]
			if partnerName == "" {
				partnerName = scenarioDiscordName(sc.Key, sc.SortLogic, *spec.DuoPartner)
			}
			duoRequest = &partnerName
		}

		_, err := seed.Pool.Exec(seed.Ctx, `
			INSERT INTO registrations (event_id, user_id, can_substitute, can_lobby_host, duo_request, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			ON CONFLICT (event_id, user_id) DO NOTHING
		`, eventID, userID, spec.CanSub, spec.CanHost, duoRequest)
		if err != nil {
			common.Fatal("failed creating registration",
				"scenario", sc.Key, "event_id", eventID, "slot", spec.Slot, "error", err)
		}
	}
	return names
}

func seedScenario(seed *common.SeedContext, hostID uuid.UUID, g *systemGames, sc Scenario, groupIndex int) SeedResult {
	groupID := scenarioGroupID(sc.Key, sc.SortLogic)
	game := g.gameForKey(sc.GameKey)
	result := SeedResult{
		Key:         sc.Key,
		SortLogic:   sc.SortLogic,
		Game:        sc.GameKey,
		GroupID:     groupID.String(),
		Description: sc.Description,
	}

	if groupExists(seed, groupID) {
		result.Skipped = true
		return result
	}

	insertEventGroup(seed, groupID, hostID, sc.SubMin, sc.Region, sc.SortLogic)

	start := firstEventStart(groupIndex)
	nextStart := start
	for i, ev := range sc.Events {
		modeName := defaultModeName(ev.ModeName)
		mode := game.modeByName(modeName)
		eventID := scenarioEventID(groupID, i)
		insertEvent(seed, eventID, groupID, mode.ID, nextStart)
		if len(ev.Registrations) > 0 {
			registerEventPlayers(seed, sc, eventID, &game, ev.Registrations)
		}
		nextStart = nextStart.Add(time.Duration(mode.Duration) * time.Minute)
	}

	slog.Info("seeded matchmaking scenario",
		"key", sc.Key, "sort_logic", sc.SortLogic, "game", sc.GameKey, "group_id", groupID)
	return result
}

func runScenarios(seed *common.SeedContext, hostName string, asJSON bool) {
	hostID := resolveHost(seed, hostName)
	games := loadSystemGames(seed)
	scenarios := allScenarios(games)

	results := make([]SeedResult, 0, len(scenarios))
	for i, sc := range scenarios {
		results = append(results, seedScenario(seed, hostID, games, sc, i))
	}

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			common.Fatal("failed encoding JSON output", "error", err)
		}
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tMODE\tGAME\tGROUP_ID\tDESCRIPTION")
	for _, r := range results {
		desc := r.Description
		if r.Skipped {
			desc = "(skipped — already exists) " + desc
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			r.Key, r.SortLogic, r.Game, r.GroupID, desc)
	}
	_ = w.Flush()

	created := 0
	skipped := 0
	for _, r := range results {
		if r.Skipped {
			skipped++
		} else {
			created++
		}
	}
	slog.Info("matchmaking seed complete", "created", created, "skipped", skipped, "total", len(results))
}

func countExpectedScenarios(g *systemGames) int {
	return len(allScenarios(g))
}

func countExistingScenarioGroups(seed *common.SeedContext, g *systemGames) int {
	count := 0
	for _, sc := range allScenarios(g) {
		if groupExists(seed, scenarioGroupID(sc.Key, sc.SortLogic)) {
			count++
		}
	}
	return count
}

func partialSeedStateError(seed *common.SeedContext, g *systemGames) {
	expected := countExpectedScenarios(g)
	found := countExistingScenarioGroups(seed, g)
	if found > 0 && found < expected {
		common.Fatal("partial matchmaking seed state detected; run cleanup before re-seeding",
			"found_groups", found, "expected_groups", expected)
	}
}
