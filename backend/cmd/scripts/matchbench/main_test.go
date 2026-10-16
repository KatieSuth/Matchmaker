package main_test

import (
	"math/rand/v2"
	"testing"
	"time"

	matchbench "github.com/KatieSuth/MatchmakerAPI/cmd/scripts/matchbench"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSizes(t *testing.T) {
	t.Parallel()

	got, err := matchbench.ParseSizes("10, 50,100")
	require.NoError(t, err)
	assert.Equal(t, []int{10, 50, 100}, got)

	for _, bad := range []string{"", "10,,50", "0", "-5", "abc"} {
		_, err := matchbench.ParseSizes(bad)
		assert.Error(t, err, bad)
	}
}

func TestParseOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		opts, err := matchbench.ParseOptions(nil)
		require.NoError(t, err)
		assert.Equal(t, []int{10, 50, 100}, opts.Sizes)
		assert.Equal(t, 1000, opts.Iterations)
		assert.Equal(t, uint64(1), opts.Seed)
	})

	t.Run("rejects invalid values", func(t *testing.T) {
		for _, args := range [][]string{
			{"-iterations=0"},
			{"-warmup=-1"},
			{"-sub-min=-1"},
			{"-sub-rate=1.5"},
			{"-duo-rate=-0.1"},
			{"-sizes=0"},
		} {
			_, err := matchbench.ParseOptions(args)
			assert.Error(t, err, args)
		}
	})
}

func TestPercentile(t *testing.T) {
	t.Parallel()

	assert.Equal(t, time.Duration(0), matchbench.Percentile(nil, 0.95))

	sorted := make([]time.Duration, 100)
	for i := range sorted {
		sorted[i] = time.Duration(i + 1)
	}
	assert.Equal(t, time.Duration(50), matchbench.Percentile(sorted, 0.50))
	assert.Equal(t, time.Duration(95), matchbench.Percentile(sorted, 0.95))
	assert.Equal(t, time.Duration(100), matchbench.Percentile(sorted, 1.0))
	assert.Equal(t, time.Duration(1), matchbench.Percentile(sorted, 0.001))
	assert.Equal(t, time.Duration(7), matchbench.Percentile([]time.Duration{7}, 0.95))
}

func TestWeightedTierInRange(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewPCG(1, 2))
	seen := map[int]bool{}
	for i := 0; i < 10000; i++ {
		tier := matchbench.WeightedTier(rng)
		require.GreaterOrEqual(t, tier, 1)
		require.LessOrEqual(t, tier, matchbench.TierWeightsCount)
		seen[tier] = true
	}
	assert.Greater(t, len(seen), 20, "ranks should span most of the ladder")
}

func TestGeneratePlayers(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewPCG(1, 2))
	players := matchbench.GeneratePlayers(rng, 100, 0.3, 0.2)
	require.Len(t, players, 100)

	byName := map[string]int{}
	for i, p := range players {
		byName[p.DiscordName] = i
	}
	duos := 0
	for _, p := range players {
		if p.DuoRequest == nil {
			continue
		}
		duos++
		partner := players[byName[*p.DuoRequest]]
		require.NotNil(t, partner.DuoRequest)
		assert.Equal(t, p.DiscordName, *partner.DuoRequest, "duo requests must be mutual")
	}
	assert.Equal(t, 20, duos)

	t.Run("deterministic ranks for a seed", func(t *testing.T) {
		a := matchbench.GeneratePlayers(rand.New(rand.NewPCG(9, 9)), 50, 0.3, 0.2)
		b := matchbench.GeneratePlayers(rand.New(rand.NewPCG(9, 9)), 50, 0.3, 0.2)
		for i := range a {
			assert.Equal(t, a[i].UserID, b[i].UserID)
			assert.Equal(t, a[i].AvgRank, b[i].AvgRank)
			assert.Equal(t, a[i].CanSubstitute, b[i].CanSubstitute)
		}
	})

	t.Run("zero rates", func(t *testing.T) {
		ps := matchbench.GeneratePlayers(rand.New(rand.NewPCG(1, 1)), 10, 0, 0)
		for _, p := range ps {
			assert.False(t, p.CanSubstitute)
			assert.Nil(t, p.DuoRequest)
		}
	})
}

func TestRun(t *testing.T) {
	t.Parallel()

	opts, err := matchbench.ParseOptions([]string{"-sizes=10,50", "-iterations=5", "-warmup=1"})
	require.NoError(t, err)

	results, err := matchbench.Run(opts)
	require.NoError(t, err)
	require.Len(t, results, 4)

	for _, r := range results {
		assert.Equal(t, 5, r.Iterations)
		assert.Contains(t, []string{"balanced", "ranked"}, r.SortLogic)
		assert.LessOrEqual(t, r.P50Micros, r.P95Micros)
		assert.LessOrEqual(t, r.P95Micros, r.P99Micros)
		assert.LessOrEqual(t, r.P99Micros, r.MaxMicros)
	}
	assert.Equal(t, 1.0, results[0].Lobbies, "10 players fill exactly one 5v5 lobby")

	t.Run("too few players errors", func(t *testing.T) {
		small, err := matchbench.ParseOptions([]string{"-sizes=5", "-iterations=1", "-warmup=0"})
		require.NoError(t, err)
		_, err = matchbench.Run(small)
		assert.Error(t, err)
	})
}
