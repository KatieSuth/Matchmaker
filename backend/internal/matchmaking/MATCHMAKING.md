# Matchmaking

This package implements team and lobby assignment for a single **game** (`events` row) when the host locks in an event group. It is pure logic: no database access. The store layer maps registrations into `Player` values, calls `PlanEvent`, and persists the result.

Matchmaking runs **independently per game** in a group. Group-level settings (`sort_logic`, `sub_min`) apply uniformly; `team_size` comes from each game's mode.

---

## Entry point

`PlanEvent(players, cfg, settings)` in `plan.go` orchestrates the full pipeline:

```
ValidateCapacity
  → strategy (balanced | ranked)
  → ApplySubCapacityRosterConstraint   (n ≥ 2 only)
  → SplitIntoTeams                     (per lobby)
  → AssignMandatorySubs                (n ≥ 2 only)
  → AssignRemainingAsSubs
  → PickLobbyHost                      (per lobby)
  → IsLobbyUnfair                      (per lobby)
```

The only difference between modes is the **strategy** step (roster pool selection and lobby distribution). All subsequent steps are shared.

---

## Input types

### `Player`

Built by `mapRegistrationsToPlayers` in `store/teams_planning.go` from `GetMatchmakingRegistrationsForEvent`:

| Field | Source | Used for |
|-------|--------|----------|
| `AvgRank` | `(current_rank_order + peak_rank_order) / 2` | All skill comparisons |
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

`AverageRankOrder(current, peak)` in `rank.go`:

```
AvgRank = (current_rank_order + peak_rank_order) / 2.0
```

Both orders must be present on the registration's `user_games` row. This single float drives sorting, roster selection, team snake drafts, and fairness checks.

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

`ValidateCapacity` returns the lobby count or a `ValidationError` with a game-specific message. Zero registrations → zero lobbies (game skipped).

### Example

5v5 (`slots = 10`), `sub_min = 3`:

- 15 registrants → **1 lobby** (need 26 for 2).
- 26 registrants, 6 sub-eligible → **2 lobbies** (20 roster + 6 mandatory subs).

---

## Mode: Balanced (`sort_logic = "balanced"`)

**Goal:** Mix skill across lobbies and form even teams. Best for casual play.

Implemented in `balanced.go` and `roster.go`.

### Step 1 — Roster pool (`selectBalancedRosterPool`)

If `needed ≥ len(players)`, everyone is rostered.

When trimming is required (`needed < len(players)`), players are picked in a repeating **high → mid → low** cycle from the remaining pool:

| Phase | Selection |
|-------|-----------|
| High | Highest `AvgRank` in remaining (`index 0` after desc sort) |
| Mid | Player closest to the remaining pool mean `AvgRank` (ties → `CompareByRankThenAvailability`) |
| Low | Lowest `AvgRank` in remaining (`index n-1`) |

`can_substitute` does **not** affect roster pool selection.

**Example:** 5 players with skills 24, 16, 10, 6, 4 — need 3:

| Pick | Phase | Skill |
|------|-------|-------|
| 1 | High | 24 |
| 2 | Mid | 10 |
| 3 | Low | 4 |

Rank 16 is excluded, not rank 10.

### Step 2 — Lobby distribution (`AssignBalanced`)

The roster pool is sorted by skill descending, then assigned via **snake draft** across lobbies (`snakeLobbyIndex`):

```
Round 0: L0, L1, L2, ...
Round 1: L2, L1, L0, ...  (reversed)
```

With one lobby, all picks go to that lobby in skill order.

**Example:** 8 players, 2 lobbies × 4 slots — lobby 0 gets skills 20, 17, 16, 13; lobby 1 gets 19, 18, 15, 14.

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
- If rank-based rostering exceeds that cap, lowest-ranked rostered sub-eligible players are swapped out for the best available non-sub players (`bestNonSubReplacement`).

**Single-lobby games (`n = 1`):** This exception does not run. `sub_min` mandatory subs also do not run.

---

## Shared post-processing

### Team split (`SplitIntoTeams`)

Within each lobby, roster players are sorted by skill descending and split into Team 1 and Team 2 via a **snake draft** (same alternating pattern as lobby snake). Each player gets `team_number = 1` or `2`.

Subs are not assigned a team number.

### Mandatory subs (`AssignMandatorySubs`)

Only when `n ≥ 2` and `sub_min > 0`.

Assigns `sub_min` sub-eligible players per lobby from unassigned players, highest skill first. Sets `team_number = nil`.

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

Detects one player far above (or below) the rest — e.g. one Radiant among Golds in a Valorant lobby.

### Check 2: Team separation (`teamSeparationExceeds`)

For roster players with `team_number` 1 or 2 (subs excluded):

```
team_avg = sum(AvgRank for team) / count(team)
unfair if |team1_avg − team2_avg| > scaledTeamSeparation
```

Detects residual imbalance after the snake team split.

### Tier scaling (`ScaledFairnessThresholds`)

Env baselines are calibrated for a reference tier count (default 25 = Valorant). They scale linearly to each game's actual tier count:

```
scale = tierCount / FairnessReferenceTierCount

scaledOutlierGap     = FairnessOutlierGap     × scale
scaledTeamSeparation = FairnessTeamSeparation × scale
```

If `tierCount ≤ 0` or `FairnessReferenceTierCount ≤ 0`, baselines are used unscaled.

| Game tiers    | Scaled outlier gap (defaults 8 @ 25) | Scaled team separation (defaults 4 @ 25) |
|---------------|--------------------------------------|------------------------------------------|
| 10            | 3.2                                  | 1.6                                      |
| 25 (Valorant) | 8.0                                  | 4.0                                      |
| 31 (LoL)      | 9.9                                  | 5.0                                      |


Tier scaling ensures fairness calculations are accurate when the amount of tiers are unknown in advance, such as for a user-defined game.

A warning means the algorithm did its best but could not achieve a tight balance — not a hard failure. The host may edit teams manually.

---

## Fairness environment variables

Parsed in `cmd/server/main.go`, passed through `handler.New` → `store.CreateTeamsForGroup` → `PlanEvent` / `IsLobbyUnfair`. Never read via `os.Getenv` inside this package.

| Variable                        | Default | Type  | Role                                                                                          |
|---------------------------------|---------|-------|-----------------------------------------------------------------------------------------------|
| `FAIRNESS_OUTLIER_GAP`          | `8`     | int   | Baseline max gap between 1st and 2nd highest skill in a lobby roster, at reference tier count |
| `FAIRNESS_TEAM_SEPARATION`      | `4`     | float | Baseline max allowed difference between team average skills, at reference tier count          |
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
```

`FairnessOutlierGap` and `FairnessTeamSeparation` are **baselines**, not absolute thresholds for every game. A 10-tier custom game uses `scale = 10/25 = 0.4`, so default outlier gap becomes `8 × 0.4 = 3.2` rank-order units.

### Tuning guidance

- **Increase** `FAIRNESS_OUTLIER_GAP` or `FAIRNESS_TEAM_SEPARATION` → fewer warnings (more tolerant of spread).
- **Decrease** → more warnings (stricter balance expectations).
- Change `FAIRNESS_REFERENCE_TIER_COUNT` only if you recalibrate baselines for a different reference ladder; it does not change a specific game's tier count (that always comes from `game_ranks`).

---

## Source file map

| File          | Responsibility                              |
|---------------|---------------------------------------------|
| `plan.go`     | `PlanEvent` orchestrator                    |
| `balanced.go` | Balanced lobby snake draft                  |
| `ranked.go`   | Ranked sequential lobby packing             |
| `roster.go`   | Balanced/ranked roster pool selection       |
| `capacity.go` | Lobby count math                            |
| `subs.go`     | Sub-capacity swaps, mandatory/overflow subs |
| `teams.go`    | Within-lobby team snake draft               |
| `tiebreak.go` | `CompareByRankThenAvailability`             |
| `rank.go`     | `AverageRankOrder`                          |
| `host.go`     | `PickLobbyHost`                             |
| `fairness.go` | Threshold scaling and `IsLobbyUnfair`       |
| `types.go`    | Domain types                                |
| `errors.go`   | `ValidationError`                           |

Tests live in `*_test.go` alongside each file (`package matchmaking_test`).

---

## Persistence (store layer)

`CreateTeamsForGroup` calls `planTeamsForGroup` (validation only, no writes), then persists inside a transaction:

- Closes registration on the group.
- Creates `lobbies` rows with `fairness_warning`.
- Inserts `players` rows with `team_number` 1, 2, or `NULL` (subs).
- Does not insert `players` rows for unplaced non-subs.

See `store/teams_planning.go` and `store/event_lobbies.go` for DB integration.
