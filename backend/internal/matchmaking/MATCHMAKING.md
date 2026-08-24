# Matchmaking

This package implements team and lobby assignment for a single **game** (`events` row) when the host locks in an event group. It is pure logic: no database access. The store layer maps registrations into `Player` values, calls `PlanEvent`, and persists the result.

Matchmaking runs **independently per game** in a group. Group-level settings (`sort_logic`, `sub_min`) apply uniformly; `team_size` comes from each game's mode.

---

## Entry point

`PlanEvent(players, cfg, settings)` in `plan.go` orchestrates the full pipeline:

```
ValidateCapacity
  → strategy (balanced window fill | ranked packing)
  → ApplySubCapacityRosterConstraint   (n ≥ 2 only)
  → ApplyDuoLobbyGrouping              (n ≥ 2 only; best-effort)
  → team split                         (windows when balanced; snake when ranked; duo pass)
  → repairUnfairWindowPairs            (balanced only; leftover in-place, adjacent fringe, then same-window combo)
  → AssignMandatorySubs                (n ≥ 2 only)
  → AssignRemainingAsSubs
  → PickLobbyHost                      (per lobby)
  → IsLobbyUnfair                      (per lobby)
  → drop n and replan                  (only if this n benches every can-sub and a lobby is still unfair)
  → clear SubCapacityAdjusted          (if every lobby is fair)
```

Modes differ in **lobby fill** and **team split**. Subs, duo passes, hosts, and fairness are shared.

---

## Input types

### `Player`

Built by `mapRegistrationsToPlayers` in `store/teams_planning.go` from `GetMatchmakingRegistrationsForEvent`:

| Field | Source | Used for |
|-------|--------|----------|
| `DiscordName` | `users.discord_name` | Mutual duo matching |
| `DuoRequest` | `registrations.duo_request` | Mutual duo matching |
| `AvgRank` | `game_ranks.order` for stored `user_games.avg_rank` | All skill comparisons |
| `CanSubstitute` | `registrations.can_substitute` | Sub pool eligibility; roster cap when n ≥ 2 |
| `CanLobbyHost` | `registrations.can_lobby_host` | Lobby host selection |
| `RegisteredGameCount` | Registrations across all games in the group | Tie-break |
| `CreatedAt` | Registration timestamp | Tie-break; host fallback |

Rank orders come from `game_ranks.order` for the event's game. Tier count for fairness scaling is `MAX(game_ranks.order)` for that game.

### `Config` (per game)

| Field | Source |
|-------|--------|
| `TeamSize` | `game_modes.team_size` |
| `Slots` | `2 × TeamSize` (roster size per lobby) |
| `SubMin` | `event_groups.sub_min` |
| `SortLogic` | `event_groups.sort_logic` (`balanced` or `ranked`) |
| `TierCount` | `MAX(game_ranks.order)` for the game |
| `LobbyCount` | Set by `ValidateCapacity` |

### `Settings` (env-backed fairness)

Injected from `main.go` → `handler.New` → `CreateTeamsForGroup`. See [Fairness environment variables](#fairness-environment-variables).

---

## Skill value

`user_games.avg_rank` stores a `game_ranks` row whose order is the floored average of current and peak:

```
floored_order = floor((current_rank_order + peak_rank_order) / 2.0)
AvgRank = game_ranks.order for user_games.avg_rank
```

`AverageRankOrder` and `FlooredAverageRankOrder` in `rank.go` compute the numeric average and floored order when persisting or backfilling. Both current and peak ranks must be present on the `user_games` row. `UpsertGameForUser` writes `avg_rank`; matchmaking backfills any missing values before planning. This single value drives sorting, roster selection, team assignment, and fairness checks.

---

## Tie-break: `CompareByRankThenAvailability`

When two players have the same `AvgRank` and a choice must be made:

1. **Higher** `AvgRank` wins (when sorting descending; inverted when sorting ascending).
2. If equal → **fewer** `RegisteredGameCount` wins (player signed up for fewer games in the group).
3. If still equal → **earlier** `CreatedAt` wins.
4. `CanSubstitute` does **not** affect this comparator.

---

## Capacity

Variables for one game:

- `slots = 2 × team_size` — roster players per lobby (one team on each side).
- `n` — number of lobbies to create.
- `sub_min` — mandatory subs **per lobby** when `n ≥ 2`.

### Minimum registrants

| Lobbies (`n`) | `required(n)` |
|---------------|---------------|
| 1 | `slots` |
| 2+ | `n × slots + n × sub_min` |

### Max lobbies (`MaxLobbies`)

Greedy: increment `n` while:

1. `registered_count ≥ required(n)`
2. If `n ≥ 2`: `substitute_count ≥ n × sub_min`

That ceiling **may bench every can-sub** (`MaxSubstituteEligibleOnRoster == 0`). `PlanEvent` still tries that `n`. If every lobby is fair after repair, it keeps the extra lobby. If any lobby is still unfair, it drops to `n − 1` and replans (and repeats while the smaller `n` also benches every can-sub). Spare can-subs on a roster (`MaxSubstituteEligibleOnRoster > 0`) stop the drop — further imbalance is left to repair and the host warning.

`ValidateCapacity` returns the lobby count or a `ValidationError` with a game-specific message. Zero registrations → zero lobbies (game skipped).

### Example

5v5 (`slots = 10`), `sub_min = 3`:

- 15 registrants → **1 lobby** (need 26 for 2).
- 26 registrants, 6 sub-eligible → **2 lobbies** if the 20 non-sub roster spots stay even; **1 lobby** if they do not (2 would bench all 6 can-subs).
- 26 registrants, 7 sub-eligible → **2 lobbies** (20 roster + 6 mandatory subs, one can-sub may still be rostered).

2v2 (`slots = 4`), `sub_min = 2`, 18 registrants, 6 sub-eligible → **3 lobbies** when the 12 non-subs form even windows; **2 lobbies** when they cannot.

---

## Mode: Balanced (`sort_logic = "balanced"`)

**Goal:** Mix skill across lobbies and form even teams. Best for casual play.

Implemented in `balanced.go`, `windows.go`, and `teams.go`. Ranked packing is a separate mode and is not used here.

### Rank windows

The game ladder (`Config.TierCount` = `MAX(game_ranks.order)`) is split into `min(teamSize, tierCount)` contiguous windows. Remainder ranks go to the **lowest** windows.

**Ideal:** each team takes **one player from each window** (two from the window, opposite sides). A 5v5 is five mirrored pairs; a 3v3 is three.

Valorant examples (`tierCount = 25`):

| Mode | Windows (order ranges) |
|------|------------------------|
| 5v5 | 1–5, 6–10, 11–15, 16–20, 21–25 |
| 3v3 | 1–9, 10–17, 18–25 |
| 2v2 | 1–13, 14–25 |

A 4-rank custom 5v5 uses 4 windows, then takes any leftover pair from the fullest band.

`AvgRank` is clamped into `1..tierCount`. If `teamSize` or `tierCount` is invalid, lobby fill and team split fall back to snake so planning cannot panic.

### Lobby fill (`AssignBalanced`)

Lobbies are dealt in **parallel**: each lobby gets its 2 from a window before any lobby gets 4 from that window.

1. Start every window at quota 2 (one player per team).
2. A window with fewer than 2 players is skipped in this pass (the singleton stays available for extras).
3. Short lobbies then take extra **pairs** from the current fullest remaining window, one pair per lobby per round, so lobby 0 cannot drain a band before lobby 1 is dealt. On a count tie, prefer the **lower-rank** window.
4. If a donor has fewer than a pair, take everyone they have and continue to the next-fullest. A leftover singleton is used only when a lobby is still short.
5. When a window is overfull, take the quota players closest to that window’s **ladder midpoint** (ties → `CompareByRankThenAvailability`).

`can_substitute` does **not** affect selection.

**Example:** 2v2 with Radiant, several Ascendants, Golds, Silver, Bronze, and Iron — one 4-player lobby. High-window midpoint ~19.5 takes two Ascendants; low-window midpoint ~7 takes Silver and Bronze. Radiant and Iron sit.

### Team split (`splitIntoTeamsWindowed`)

After shared lobby post-passes, `splitIntoTeamsWindowed` runs `windowDraftIntoTeams` then the same “balance wins” duo post-pass as ranked mode. The roster is bucketed into the same windows:

- A clean pair in a window: opposite teams. The higher of the pair goes to the team with the **lower rank sum so far**.
- Four from one window (fallback): pair neighbors (1st vs 2nd, 3rd vs 4th) the same way.
- Odd leftovers: pair leftovers closest in rank first (adjacent windows win equal distance). Do not attach an outlier to a far leftover who has a closer counterpart. The last unpaired leftover goes to the weaker team.

---

## Mode: Rank Grouping (`sort_logic = "ranked"`)

**Goal:** Keep similar skill in the same lobby. Best for serious practice.

Implemented in `ranked.go` and `roster.go`.

### Step 1 — Roster pool (`selectRankedRosterPool`)

If everyone fits, all players are rostered (sorted ascending).

When trimming is required, the **majority skill side** of the pool is kept:

1. Sort ascending by `AvgRank`.
2. Compute pool mean skill.
3. Count players strictly below vs. strictly above the mean.
4. More below → keep lowest `needed` players (`sorted[:needed]`).
5. More above → keep highest `needed` players (`sorted[len-needed:]`).
6. Tie → keep the low side.

`can_substitute` does **not** affect roster pool selection.

### Step 2 — Lobby distribution (`AssignRanked`)

The roster pool (ascending) is packed **sequentially** into lobbies — no snake:

```
Lobby 0: players [0 .. slots-1]
Lobby 1: players [slots .. 2×slots-1]
...
```

**Example:** 8 players, skills 2–5 and 17–20, 2 lobbies × 4:

- Lobby 0: 2, 3, 4, 5
- Lobby 1: 17, 18, 19, 20

---

## `can_substitute` and roster priority

**Default rule:** `can_substitute` does **not** affect roster selection or tie-breaks.

**Exception (multi-lobby only):** When `n ≥ 2`, `ApplySubCapacityRosterConstraint` ensures enough sub-eligible players remain unrostered for mandatory subs (`n × sub_min`).

- Max sub-eligible on rosters = `substitute_count − (n × sub_min)` (see `MaxSubstituteEligibleOnRoster`).
- If rank-based rostering exceeds that cap, lowest-ranked rostered sub-eligible players are swapped out for an unrostered non-sub in the **same rank window** (`bestNonSubReplacement`). If that window has no eligible non-sub, the nearest window is used; within a band, pick closest to the vacated rank (ties → `CompareByRankThenAvailability`). Do not insert the globally highest leftover — that is how a Radiant lands in a Gold/Silver hole and turns a 2+2 (or 2-per-band 5v5) into 3+1.
- When any of those swaps run **and at least one lobby is still unfair** after repair and the rest of the pipeline, `GamePlan.SubCapacityAdjusted` is true so the host can be told the substitute minimum changed the lineup. `PlanEvent` clears the flag via `anyLobbyUnfair` when every lobby ends fair.

**Single-lobby games (`n = 1`):** This exception does not run. `sub_min` mandatory subs also do not run.

---

## Shared post-processing

Lobby sub-capacity swaps, duo lobby grouping, subs, hosts, and fairness flags are shared. Team split uses windows when `sort_logic` is balanced and snake when ranked; both still run the duo team pass. Balanced mode then runs leftover fairness repair when a lobby is unfair; ranked does not.

### Duo requests (`ApplyDuoLobbyGrouping` + team pass in `SplitIntoTeams`)

Players may list a **duo request** (Discord name) at registration. Matchmaking attempts to honor **mutual** pairs — both players must list each other's Discord name (case-insensitive, trimmed).

**Balance always wins.** Duo swaps are only applied when they do not worsen the baseline balance metric from the preceding step:

| Stage | When | Baseline metric |
|-------|------|-----------------|
| Lobby grouping | After sub-capacity constraint; skipped when `n = 1` | Cross-lobby average spread (`max lobby avg − min lobby avg`) |
| Team split | After window pairs (balanced) or snake draft (ranked) | Team average separation (`|avg(team1) − avg(team2)|`) |

Rules:

- Invalid or one-sided requests are ignored.
- At most **one honored duo per team** (team split only).
- Multiple mutual pairs may share a lobby; there is no per-lobby duo cap.
- Partners who are not both rostered (sub or unplaced) cannot be united.
- Duo placement is best-effort; fairness may require partners to stay in different lobbies or on opposite teams.

### Team split

**Ranked** (`SplitIntoTeams`): roster sorted by skill descending, snake draft into Team 1 and Team 2 (1–2–2–1), then duo post-pass against that baseline.

**Balanced** (`splitIntoTeamsWindowed`): `windowDraftIntoTeams` pairing as above, then the same duo post-pass against the window-split baseline.

Subs are not assigned a team number.

### Fairness repair (`repairUnfairWindowPairs`)

Balanced only, after team split and before mandatory subs. Fill and sub-capacity never look back at unplaced players; a lobby can end up **Plat vs Bronze** while a **Gold 2** sits unplaced even though swapping them in place would even the sides.

The pass runs while `IsLobbyUnfair` is true — **either** `teamSeparationExceeds` **or** `outlierExceeds`. Other lobbies are never raided (leftovers are unplaced players only; anyone already on a roster or sub pool is ineligible). Unplaced can-subs are eligible only when enough can-subs would remain unrostered for `n × sub_min`. Apply a change only when the failing check **strictly improves**, and do not newly trip the other check (an outlier fix must not create a team-separation warning, and a team-sep fix must not create an outlier warning). If both checks fail, team-average separation is optimized first. One improving change per pass; repeat while the lobby is still unfair and a swap still helps, up to `min(window count + 2, 8)` passes.

Each pass tries, in order:

1. **In-place 1-for-1, same window.** Replace one rostered player with one unplaced leftover from that rank band. Everyone else keeps their `team_number`. This is the host-style fix (Plat 1 off, leftover Gold 2 on that seat) and does not re-deal the lobby. The leftover must not widen that window's rank spread (a leftover Radiant cannot replace an Ascendant to paper over Plat vs Bronze). Among accepted swaps, pick the best primary metric, then the closer rank delta; on a sep+delta tie, replace the lower-ranked seat so a high leftover cannot mask a low-window hole.
2. **In-place 1-for-1, adjacent fringe.** Same seat-preserving swap, but the leftover may come from a neighboring window if **both** ranks sit within 2 ladder steps of the shared edge (e.g. 5v5 Gold 1 ↔ Plat 1). A Radiant is not “adjacent” to Gold.
3. **Same-window combo re-split.** If no 1-for-1 helps, the `k` rostered players in a window plus unplaced players in that window are retried as combinations (current occupants always stay in the search pool; leftovers may be trimmed so the combination count stays small) and the **whole lobby** is re-split with `splitIntoTeamsWindowed` (window pairing + duo pass). Combinations that widen that window's rank spread are skipped. Ties keep more of the current window roster (less churn).

This does **not** set `SubCapacityAdjusted`. If repair (or the rest of the pipeline) leaves every lobby fair, `PlanEvent` clears that flag so the host is not warned. Ranked mode never runs it.

### Mandatory subs (`AssignMandatorySubs`)

Only when `n ≥ 2` and `sub_min > 0`.

Assigns `sub_min` sub-eligible players per lobby from unassigned sub-eligible players, highest skill first. Sets `team_number = nil`.

### Overflow subs (`AssignRemainingAsSubs`)

Remaining sub-eligible players not on a team are placed in lobby sub pools (`team_number = nil`), distributed **round-robin** across lobbies (highest skill first in the overflow list).

### Unplaced players

`can_substitute = false` players not on a roster receive **no** `players` row. They remain registered but unplaced for that game.

| State | `players` row | `team_number` | `can_substitute` |
|-------|---------------|---------------|------------------|
| On team | yes | 1 or 2 | either |
| Sub | yes | `NULL` | must be `true` |
| Unplaced | no | — | must be `false` |

### Lobby host (`PickLobbyHost`)

Per lobby, among roster + subs:

1. First `can_lobby_host = true` by earliest `CreatedAt`.
2. Otherwise first assigned player by earliest `CreatedAt`.

---

## Fairness determination

Fairness is evaluated **per lobby** after all assignment. The result is stored as `lobbies.fairness_warning` at lock-in and returned in the API. It is not recomputed on page load.

`IsLobbyUnfair(lobby, settings, tierCount)` returns true if **either** check fails.

### Check 1: Outlier (`outlierExceeds`)

Sort roster `AvgRank` values ascending. Let `highest` and `second` be the top two.

```
unfair if (highest − second) > scaledOutlierGap
```

Detects one player far above the rest — e.g. one Platinum among Bronzes in a 2v2. Balanced leftover repair also tries to shrink this gap.

### Check 2: Team separation (`teamSeparationExceeds`)

For roster players with `team_number` 1 or 2 (subs excluded):

```
team_avg = sum(AvgRank for team) / count(team)
unfair if |team1_avg − team2_avg| > scaledTeamSeparation
```

Detects residual imbalance after the team split (and after balanced fairness repair).

### Tier scaling (`ScaledFairnessThresholds`)

Env baselines are calibrated for a reference tier count (default 25 = Valorant). They scale linearly to each game's actual tier count:

```
scale = tierCount / FairnessReferenceTierCount

scaledOutlierGap     = FairnessOutlierGap     × scale
scaledTeamSeparation = FairnessTeamSeparation × scale
```

If `tierCount ≤ 0` or `FairnessReferenceTierCount ≤ 0`, baselines are used unscaled.

| Game tiers    | Scaled outlier gap (defaults 6 @ 25) | Scaled team separation (defaults 3 @ 25) |
|---------------|--------------------------------------|------------------------------------------|
| 10            | 2.4                                  | 1.2                                      |
| 25 (Valorant) | 6.0                                  | 3.0                                      |
| 31 (LoL)      | 7.4                                  | 3.7                                      |


Tier scaling ensures fairness calculations are accurate when the amount of tiers are unknown in advance, such as for a user-defined game.

A warning means the algorithm did its best but could not achieve a tight balance — not a hard failure. The host may edit teams manually.

---

## Fairness environment variables

Parsed in `cmd/server/main.go`, passed through `handler.New` → `store.CreateTeamsForGroup` → `PlanEvent` / `IsLobbyUnfair`. Never read via `os.Getenv` inside this package.

| Variable                        | Default | Type  | Role                                                                                          |
|---------------------------------|---------|-------|-----------------------------------------------------------------------------------------------|
| `FAIRNESS_OUTLIER_GAP`          | `6`     | int   | Baseline max gap between 1st and 2nd highest skill in a lobby roster, at reference tier count |
| `FAIRNESS_TEAM_SEPARATION`      | `3`     | float | Baseline max allowed difference between team average skills, at reference tier count          |
| `FAIRNESS_REFERENCE_TIER_COUNT` | `25`    | int   | Tier count the baseline values were calibrated for                                            |

All three must be positive if set; invalid values cause startup failure.

### How they flow

```
.env
  → main.go (parse + validate)
  → matchmaking.Settings on handler.Handler
  → CreateTeamsForGroup(..., settings)
  → planTeamsForGroup: GetMaxRankOrderForGame → tierCount per game
  → PlanEvent(players, cfg, settings)
  → IsLobbyUnfair(lobby, settings, cfg.TierCount) per lobby
  → clear SubCapacityAdjusted if !anyLobbyUnfair
```

`FairnessOutlierGap` and `FairnessTeamSeparation` are **baselines**, not absolute thresholds for every game. A 10-tier custom game uses `scale = 10/25 = 0.4`, so default outlier gap becomes `6 × 0.4 = 2.4` rank-order units.

### Tuning guidance

- **Increase** `FAIRNESS_OUTLIER_GAP` or `FAIRNESS_TEAM_SEPARATION` → fewer warnings (more tolerant of spread).
- **Decrease** → more warnings (stricter balance expectations).
- Change `FAIRNESS_REFERENCE_TIER_COUNT` only if you recalibrate baselines for a different reference ladder; it does not change a specific game's tier count (that always comes from `game_ranks`).

---

## Source file map

| File          | Responsibility                                      |
|---------------|-----------------------------------------------------|
| `duos.go`     | Mutual duo detection, lobby grouping, team grouping |
| `plan.go`     | `PlanEvent` orchestrator; may drop `n` via `shouldDropReservedSubLobby`; clears `SubCapacityAdjusted` via `anyLobbyUnfair` |
| `balanced.go` | Balanced lobby fill from rank windows                 |
| `windows.go`  | Rank windows, midpoint picks, windowed team pairing   |
| `ranked.go`   | Ranked sequential lobby packing                       |
| `roster.go`   | Ranked roster pool selection                          |
| `capacity.go` | Lobby count math                                      |
| `subs.go`     | Sub-capacity swaps, mandatory/overflow subs           |
| `teams.go`    | Ranked snake draft (`SplitIntoTeams`); snake fallback |
| `tiebreak.go` | `CompareByRankThenAvailability`                     |
| `rank.go`     | `AverageRankOrder`, `FlooredAverageRankOrder`       |
| `host.go`     | `PickLobbyHost`                                     |
| `fairness.go` | Threshold scaling and `IsLobbyUnfair`               |
| `repair.go`   | Balanced leftover in-place, fringe, and combo repair |
| `types.go`    | Domain types                                        |
| `errors.go`   | `ValidationError`                                   |

Tests live in `*_test.go` alongside each file (`package matchmaking_test`).

---

## Persistence (store layer)

`CreateTeamsForGroup` (`store/events.go`) calls `planTeamsForGroup` (full `PlanEvent` per game, no writes) so a validation failure cannot close registration, then persists inside a transaction:

- Closes registration on the group.
- Creates `lobbies` rows with `fairness_warning`.
- Inserts `players` rows with `team_number` 1, 2, or `NULL` (subs).
- Does not insert `players` rows for unplaced non-subs.

The HTTP response is 200 with `{ "sub_capacity_adjusted": bool }`. That bool is the OR of each game's `GamePlan.SubCapacityAdjusted`: the host modal appears if **any** game in the group still has an unfair lobby after a substitute-minimum swap. A repaired, fully fair 2v2 does not hide the warning if a 5v5 in the same group stayed unfair.

See `store/teams_planning.go` for planning and `persistTeamPlans`. `store/event_lobbies.go` is the read path when loading saved lobbies.
