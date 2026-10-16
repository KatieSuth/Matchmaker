// Command matchbench measures PlanEvent latency (p50/p95/p99) for a Valorant 5v5
// across player counts and both sort logics. Pools are generated from a fixed seed,
// so repeated runs with the same flags plan identical inputs.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
)

const (
	valorantTierCount = 25
	valorantTeamSize  = 5
)

var sortLogics = []string{"balanced", "ranked"}

// valorantTierWeights approximates the Valorant ladder population per tier group
// (Iron … Immortal, three divisions each, then Radiant), peaking around Silver/Gold.
var valorantTierWeights = func() []float64 {
	groups := []float64{6, 11, 14, 15, 13, 10, 7, 4} // Iron, Bronze, Silver, Gold, Plat, Diamond, Ascendant, Immortal
	weights := make([]float64, 0, valorantTierCount)
	for _, g := range groups {
		weights = append(weights, g, g, g)
	}
	return append(weights, 0.5) // Radiant
}()

type options struct {
	Sizes      []int
	Iterations int
	Warmup     int
	Seed       uint64
	SubMin     int
	SubRate    float64
	DuoRate    float64
	JSON       bool
}

type result struct {
	SortLogic      string  `json:"sort_logic"`
	Players        int     `json:"players"`
	Iterations     int     `json:"iterations"`
	Lobbies        float64 `json:"avg_lobbies"`
	UnfairLobbyPct float64 `json:"unfair_lobby_pct"`
	P50Micros      float64 `json:"p50_us"`
	P95Micros      float64 `json:"p95_us"`
	P99Micros      float64 `json:"p99_us"`
	MaxMicros      float64 `json:"max_us"`
}

// main parses flags, runs every size × sort logic, and prints a table or JSON.
func main() {
	opts, err := parseOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	results, err := run(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if opts.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	printTable(opts, results)
}

// parseOptions reads CLI flags and rejects values that would make the benchmark meaningless.
func parseOptions(args []string) (options, error) {
	fs := flag.NewFlagSet("matchbench", flag.ContinueOnError)
	sizes := fs.String("sizes", "10,50,100", "comma-separated player counts")
	iterations := fs.Int("iterations", 1000, "timed PlanEvent calls per size and sort logic")
	warmup := fs.Int("warmup", 100, "untimed calls per size and sort logic before measuring")
	seed := fs.Uint64("seed", 1, "RNG seed for player pools")
	subMin := fs.Int("sub-min", 2, "event group sub_min")
	subRate := fs.Float64("sub-rate", 0.3, "fraction of players with can_substitute")
	duoRate := fs.Float64("duo-rate", 0.2, "fraction of players in a mutual duo request")
	asJSON := fs.Bool("json", false, "emit results as JSON")
	if err := fs.Parse(args); err != nil {
		return options{}, err
	}

	parsed, err := parseSizes(*sizes)
	if err != nil {
		return options{}, err
	}
	if *iterations < 1 || *warmup < 0 || *subMin < 0 {
		return options{}, fmt.Errorf("iterations must be >= 1; warmup and sub-min must be >= 0")
	}
	if *subRate < 0 || *subRate > 1 || *duoRate < 0 || *duoRate > 1 {
		return options{}, fmt.Errorf("sub-rate and duo-rate must be within [0, 1]")
	}

	return options{
		Sizes:      parsed,
		Iterations: *iterations,
		Warmup:     *warmup,
		Seed:       *seed,
		SubMin:     *subMin,
		SubRate:    *subRate,
		DuoRate:    *duoRate,
		JSON:       *asJSON,
	}, nil
}

// parseSizes converts "10,50,100" into positive ints.
func parseSizes(raw string) ([]int, error) {
	var sizes []int
	for _, part := range strings.Split(raw, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 1 {
			return nil, fmt.Errorf("invalid size %q", part)
		}
		sizes = append(sizes, n)
	}
	return sizes, nil
}

// run benchmarks each size against both sort logics. Each iteration plans the same
// generated pool under both logics, alternating which goes first so neither
// consistently benefits from a warm cache.
func run(opts options) ([]result, error) {
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: valorantTierCount,
	}

	var results []result
	for _, size := range opts.Sizes {
		// Seed per size so adding/removing sizes does not change other sizes' pools.
		rng := rand.New(rand.NewPCG(opts.Seed, uint64(size)))
		samples := map[string][]time.Duration{}
		lobbies := map[string]int{}
		unfair := map[string]int{}

		runtime.GC()
		for i := 0; i < opts.Warmup+opts.Iterations; i++ {
			pool := generatePlayers(rng, size, opts.SubRate, opts.DuoRate)
			timed := i >= opts.Warmup

			order := sortLogics
			if i%2 == 1 {
				order = []string{sortLogics[1], sortLogics[0]}
			}
			for _, logic := range order {
				cfg := matchmaking.Config{
					EventID:   uuid.Nil,
					TeamSize:  valorantTeamSize,
					Slots:     valorantTeamSize * 2,
					SubMin:    opts.SubMin,
					SortLogic: logic,
					TierCount: valorantTierCount,
					GameLabel: "Valorant 5v5",
				}
				players := slices.Clone(pool)

				start := time.Now()
				plan, err := matchmaking.PlanEvent(players, cfg, settings)
				elapsed := time.Since(start)
				if err != nil {
					return nil, fmt.Errorf("%s with %d players: %w", logic, size, err)
				}
				if !timed {
					continue
				}
				samples[logic] = append(samples[logic], elapsed)
				lobbies[logic] += len(plan.Lobbies)
				for _, l := range plan.Lobbies {
					if l.FairnessWarning {
						unfair[logic]++
					}
				}
			}
		}

		for _, logic := range sortLogics {
			s := samples[logic]
			slices.Sort(s)
			unfairPct := 0.0
			if lobbies[logic] > 0 {
				unfairPct = 100 * float64(unfair[logic]) / float64(lobbies[logic])
			}
			results = append(results, result{
				SortLogic:      logic,
				Players:        size,
				Iterations:     len(s),
				Lobbies:        float64(lobbies[logic]) / float64(len(s)),
				UnfairLobbyPct: unfairPct,
				P50Micros:      micros(percentile(s, 0.50)),
				P95Micros:      micros(percentile(s, 0.95)),
				P99Micros:      micros(percentile(s, 0.99)),
				MaxMicros:      micros(s[len(s)-1]),
			})
		}
	}
	return results, nil
}

// generatePlayers builds a pool of n players with ladder-weighted ranks, a subRate
// share of can-subs, and roughly duoRate of players paired into mutual duo requests.
func generatePlayers(rng *rand.Rand, n int, subRate, duoRate float64) []matchmaking.Player {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	players := make([]matchmaking.Player, n)
	for i := range players {
		players[i] = matchmaking.Player{
			UserID:              seededUUID(rng),
			DiscordName:         fmt.Sprintf("player%03d", i),
			AvgRank:             float64(weightedTier(rng)),
			CanSubstitute:       rng.Float64() < subRate,
			CanLobbyHost:        rng.Float64() < 0.5,
			RegisteredGameCount: 1 + rng.IntN(3),
			CreatedAt:           base.Add(time.Duration(i) * time.Minute),
		}
	}

	// Pair adjacent players from a shuffled index list so duo partners have unrelated ranks.
	duoPairs := int(float64(n) * duoRate / 2)
	idx := rng.Perm(n)
	for p := 0; p < duoPairs; p++ {
		a, b := &players[idx[2*p]], &players[idx[2*p+1]]
		aReq, bReq := b.DiscordName, a.DiscordName
		a.DuoRequest, b.DuoRequest = &aReq, &bReq
	}
	return players
}

// seededUUID builds a v4-shaped UUID from rng so pools are fully reproducible
// (uuid.New reads crypto/rand).
func seededUUID(rng *rand.Rand) uuid.UUID {
	var id uuid.UUID
	for i := range id {
		id[i] = byte(rng.Uint32())
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return id
}

// weightedTier draws a 1-based tier order from valorantTierWeights.
func weightedTier(rng *rand.Rand) int {
	total := 0.0
	for _, w := range valorantTierWeights {
		total += w
	}
	r := rng.Float64() * total
	for i, w := range valorantTierWeights {
		r -= w
		if r < 0 {
			return i + 1
		}
	}
	return len(valorantTierWeights)
}

// percentile returns the nearest-rank percentile (p in (0, 1]) of an ascending slice.
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	rank := int(math.Ceil(p*float64(len(sorted)))) - 1
	return sorted[max(0, min(rank, len(sorted)-1))]
}

// micros converts a duration to fractional microseconds.
func micros(d time.Duration) float64 {
	return float64(d) / float64(time.Microsecond)
}

// printTable writes a human-readable summary including the run parameters needed to reproduce it.
func printTable(opts options, results []result) {
	fmt.Printf("Valorant 5v5 PlanEvent latency — seed=%d iterations=%d warmup=%d sub_min=%d sub_rate=%.2f duo_rate=%.2f\n",
		opts.Seed, opts.Iterations, opts.Warmup, opts.SubMin, opts.SubRate, opts.DuoRate)
	fmt.Printf("%s %s/%s, GOMAXPROCS=%d\n\n", runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "players\tsort_logic\tlobbies\tunfair %\tp50 µs\tp95 µs\tp99 µs\tmax µs\t")
	for _, r := range results {
		fmt.Fprintf(w, "%d\t%s\t%.1f\t%.1f\t%.1f\t%.1f\t%.1f\t%.1f\t\n",
			r.Players, r.SortLogic, r.Lobbies, r.UnfairLobbyPct, r.P50Micros, r.P95Micros, r.P99Micros, r.MaxMicros)
	}
	w.Flush()
}
