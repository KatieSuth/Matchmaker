package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func duoRequest(name string) *string {
	return &name
}

func TestApplyDuoLobbyGrouping_SingleLobbyNoOp(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	lobbies := []matchmaking.LobbyPlan{{
		Roster: []matchmaking.Player{
			{UserID: a, DiscordName: "Alpha", DuoRequest: duoRequest("Beta"), AvgRank: 10, CreatedAt: now},
			{UserID: b, DiscordName: "Beta", DuoRequest: duoRequest("Alpha"), AvgRank: 9, CreatedAt: now},
		},
	}}

	out := matchmaking.ApplyDuoLobbyGrouping(lobbies)
	require.Len(t, out[0].Roster, 2)
	assert.Equal(t, a, out[0].Roster[0].UserID)
	assert.Equal(t, b, out[0].Roster[1].UserID)
}

func TestApplyDuoLobbyGrouping_UnitesPartnersWhenSpreadAllows(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()
	lobbies := []matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{
			{UserID: a, DiscordName: "Alpha", DuoRequest: duoRequest("Charlie"), AvgRank: 10, CreatedAt: now},
			{UserID: b, DiscordName: "Bravo", AvgRank: 9, CreatedAt: now.Add(time.Minute)},
		}},
		{Roster: []matchmaking.Player{
			{UserID: c, DiscordName: "Charlie", DuoRequest: duoRequest("Alpha"), AvgRank: 8, CreatedAt: now.Add(2 * time.Minute)},
			{UserID: d, DiscordName: "Delta", AvgRank: 7, CreatedAt: now.Add(3 * time.Minute)},
		}},
	}

	out := matchmaking.ApplyDuoLobbyGrouping(lobbies)
	lobbyFor := func(id uuid.UUID) int {
		for i, lobby := range out {
			for _, p := range lobby.Roster {
				if p.UserID == id {
					return i
				}
			}
		}
		return -1
	}
	assert.Equal(t, lobbyFor(a), lobbyFor(c))
}

func TestApplyDuoLobbyGrouping_KeepsBaselineWhenSpreadWouldWorsen(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()
	lobbies := []matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{
			{UserID: a, DiscordName: "Alpha", DuoRequest: duoRequest("Charlie"), AvgRank: 20, CreatedAt: now},
			{UserID: b, DiscordName: "Bravo", AvgRank: 5, CreatedAt: now.Add(time.Minute)},
		}},
		{Roster: []matchmaking.Player{
			{UserID: c, DiscordName: "Charlie", DuoRequest: duoRequest("Alpha"), AvgRank: 18, CreatedAt: now.Add(2 * time.Minute)},
			{UserID: d, DiscordName: "Delta", AvgRank: 4, CreatedAt: now.Add(3 * time.Minute)},
		}},
	}

	out := matchmaking.ApplyDuoLobbyGrouping(lobbies)
	assert.Equal(t, a, out[0].Roster[0].UserID)
	assert.Equal(t, b, out[0].Roster[1].UserID)
	assert.Equal(t, c, out[1].Roster[0].UserID)
	assert.Equal(t, d, out[1].Roster[1].UserID)
}

func TestApplyDuoLobbyGrouping_IgnoresOneSidedRequest(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	c := uuid.New()
	lobbies := []matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{
			{UserID: a, DiscordName: "Alpha", DuoRequest: duoRequest("Charlie"), AvgRank: 10, CreatedAt: now},
		}},
		{Roster: []matchmaking.Player{
			{UserID: c, DiscordName: "Charlie", AvgRank: 8, CreatedAt: now.Add(time.Minute)},
		}},
	}

	out := matchmaking.ApplyDuoLobbyGrouping(lobbies)
	assert.Len(t, out[0].Roster, 1)
	assert.Len(t, out[1].Roster, 1)
}

func TestApplyDuoLobbyGrouping_MatchesCaseAndWhitespace(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	c := uuid.New()
	b := uuid.New()
	d := uuid.New()
	lobbies := []matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{
			{UserID: a, DiscordName: "Alpha", DuoRequest: duoRequest("  charlie  "), AvgRank: 10, CreatedAt: now},
			{UserID: b, DiscordName: "Bravo", AvgRank: 9, CreatedAt: now.Add(time.Minute)},
		}},
		{Roster: []matchmaking.Player{
			{UserID: c, DiscordName: "CHARLIE", DuoRequest: duoRequest("alpha"), AvgRank: 8, CreatedAt: now.Add(2 * time.Minute)},
			{UserID: d, DiscordName: "Delta", AvgRank: 7, CreatedAt: now.Add(3 * time.Minute)},
		}},
	}

	out := matchmaking.ApplyDuoLobbyGrouping(lobbies)
	lobbyFor := func(id uuid.UUID) int {
		for i, lobby := range out {
			for _, p := range lobby.Roster {
				if p.UserID == id {
					return i
				}
			}
		}
		return -1
	}
	assert.Equal(t, lobbyFor(a), lobbyFor(c))
}

func TestSplitIntoTeams_KeepsMutualDuoAlreadyOnSameTeam(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()
	// Snake draft places 2nd and 3rd highest ranks on the same team; those partners stay together.
	roster := []matchmaking.Player{
		{UserID: a, DiscordName: "Alpha", AvgRank: 20, CreatedAt: now},
		{UserID: b, DiscordName: "Bravo", DuoRequest: duoRequest("Charlie"), AvgRank: 18, CreatedAt: now.Add(time.Minute)},
		{UserID: c, DiscordName: "Charlie", DuoRequest: duoRequest("Bravo"), AvgRank: 17, CreatedAt: now.Add(2 * time.Minute)},
		{UserID: d, DiscordName: "Delta", AvgRank: 10, CreatedAt: now.Add(3 * time.Minute)},
	}

	team1, team2 := matchmaking.SplitIntoTeams(roster, 2)
	teamFor := func(id uuid.UUID) int {
		for _, p := range team1 {
			if p.UserID == id {
				return 1
			}
		}
		for _, p := range team2 {
			if p.UserID == id {
				return 2
			}
		}
		return 0
	}
	assert.Equal(t, teamFor(b), teamFor(c))
}

func TestApplyDuoTeamGroupingForTest_HonorsDuoWithoutWorseningBalance(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()
	team1 := []matchmaking.Player{
		{UserID: a, DiscordName: "Alpha", DuoRequest: duoRequest("Charlie"), AvgRank: 10, CreatedAt: now},
		{UserID: b, DiscordName: "Bravo", AvgRank: 9, CreatedAt: now},
	}
	team2 := []matchmaking.Player{
		{UserID: c, DiscordName: "Charlie", DuoRequest: duoRequest("Alpha"), AvgRank: 9, CreatedAt: now},
		{UserID: d, DiscordName: "Delta", AvgRank: 8, CreatedAt: now},
	}
	baseline := matchmaking.TeamAverageSeparationForTest(team1, team2)

	team1, team2 = matchmaking.ApplyDuoTeamGroupingForTest(team1, team2, baseline)
	teamFor := func(id uuid.UUID) int {
		for _, p := range team1 {
			if p.UserID == id {
				return 1
			}
		}
		for _, p := range team2 {
			if p.UserID == id {
				return 2
			}
		}
		return 0
	}
	assert.Equal(t, 1, teamFor(a))
	assert.Equal(t, 1, teamFor(c))
	assert.LessOrEqual(t, matchmaking.TeamAverageSeparationForTest(team1, team2), baseline+balanceEpsilonForTest)
}

const balanceEpsilonForTest = 1e-9

func TestSplitIntoTeams_KeepsBalanceWhenDuoWouldWorsenSeparation(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()
	roster := []matchmaking.Player{
		{UserID: a, DiscordName: "Alpha", DuoRequest: duoRequest("Charlie"), AvgRank: 20, CreatedAt: now},
		{UserID: b, DiscordName: "Bravo", AvgRank: 10, CreatedAt: now.Add(time.Minute)},
		{UserID: c, DiscordName: "Charlie", DuoRequest: duoRequest("Alpha"), AvgRank: 18, CreatedAt: now.Add(2 * time.Minute)},
		{UserID: d, DiscordName: "Delta", AvgRank: 8, CreatedAt: now.Add(3 * time.Minute)},
	}

	team1, team2 := matchmaking.SplitIntoTeams(roster, 2)
	teamFor := func(id uuid.UUID) int {
		for _, p := range team1 {
			if p.UserID == id {
				return 1
			}
		}
		for _, p := range team2 {
			if p.UserID == id {
				return 2
			}
		}
		return 0
	}
	assert.NotEqual(t, teamFor(a), teamFor(c))
}

func TestSplitIntoTeams_OnlyOneDuoPerTeam(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()
	e := uuid.New()
	f := uuid.New()
	roster := []matchmaking.Player{
		{UserID: a, DiscordName: "A", DuoRequest: duoRequest("B"), AvgRank: 20, CreatedAt: now},
		{UserID: b, DiscordName: "B", DuoRequest: duoRequest("A"), AvgRank: 19, CreatedAt: now.Add(time.Minute)},
		{UserID: c, DiscordName: "C", DuoRequest: duoRequest("D"), AvgRank: 18, CreatedAt: now.Add(2 * time.Minute)},
		{UserID: d, DiscordName: "D", DuoRequest: duoRequest("C"), AvgRank: 17, CreatedAt: now.Add(3 * time.Minute)},
		{UserID: e, DiscordName: "E", AvgRank: 10, CreatedAt: now.Add(4 * time.Minute)},
		{UserID: f, DiscordName: "F", AvgRank: 9, CreatedAt: now.Add(5 * time.Minute)},
	}

	team1, team2 := matchmaking.SplitIntoTeams(roster, 3)
	duosOnTeam := func(team []matchmaking.Player) int {
		count := 0
		for _, pair := range []struct{ x, y uuid.UUID }{
			{a, b}, {c, d},
		} {
			onTeam := func(id uuid.UUID) bool {
				for _, p := range team {
					if p.UserID == id {
						return true
					}
				}
				return false
			}
			if onTeam(pair.x) && onTeam(pair.y) {
				count++
			}
		}
		return count
	}
	assert.LessOrEqual(t, duosOnTeam(team1), 1)
	assert.LessOrEqual(t, duosOnTeam(team2), 1)
}

func TestLobbyAverageSpreadForTest(t *testing.T) {
	spread := matchmaking.LobbyAverageSpreadForTest([]matchmaking.LobbyPlan{
		{Roster: []matchmaking.Player{{AvgRank: 10}, {AvgRank: 12}}},
		{Roster: []matchmaking.Player{{AvgRank: 6}, {AvgRank: 8}}},
	})
	assert.InDelta(t, 4, spread, 0.01)
}

func TestTeamAverageSeparationForTest(t *testing.T) {
	sep := matchmaking.TeamAverageSeparationForTest(
		[]matchmaking.Player{{AvgRank: 15}, {AvgRank: 14}},
		[]matchmaking.Player{{AvgRank: 13}, {AvgRank: 12}},
	)
	assert.InDelta(t, 2, sep, 0.01)
}

func TestFindMutualDuoPairsForTest(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	pairs := matchmaking.FindMutualDuoPairsForTest([]matchmaking.Player{
		{UserID: a, DiscordName: "Alpha", DuoRequest: duoRequest("Beta"), CreatedAt: now},
		{UserID: b, DiscordName: "Beta", DuoRequest: duoRequest("Alpha"), CreatedAt: now},
		{UserID: c, DiscordName: "Charlie", DuoRequest: duoRequest("Alpha"), CreatedAt: now},
	})
	require.Len(t, pairs, 1)
	assert.Equal(t, a, pairs[0].A.UserID)
	assert.Equal(t, b, pairs[0].B.UserID)
}

func TestSortMutualDuoPairsForTest_OrdersDeterministically(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()
	pairs := matchmaking.SortMutualDuoPairsForTest([]matchmaking.Player{
		{UserID: c, DiscordName: "C", DuoRequest: duoRequest("D"), CreatedAt: now.Add(2 * time.Minute)},
		{UserID: d, DiscordName: "D", DuoRequest: duoRequest("C"), CreatedAt: now.Add(3 * time.Minute)},
		{UserID: a, DiscordName: "A", DuoRequest: duoRequest("B"), CreatedAt: now},
		{UserID: b, DiscordName: "B", DuoRequest: duoRequest("A"), CreatedAt: now.Add(time.Minute)},
	})
	require.Len(t, pairs, 2)
	assert.Equal(t, a, pairs[0].A.UserID)
	assert.Equal(t, c, pairs[1].A.UserID)
}

func TestCompareDuoPairsForTest(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()

	earlierPair := matchmaking.DuoPairForTest{
		A: matchmaking.Player{UserID: a, CreatedAt: now},
		B: matchmaking.Player{UserID: b, CreatedAt: now},
	}
	laterPair := matchmaking.DuoPairForTest{
		A: matchmaking.Player{UserID: c, CreatedAt: later},
		B: matchmaking.Player{UserID: d, CreatedAt: later},
	}
	assert.Less(t, matchmaking.CompareDuoPairsForTest(earlierPair, laterPair), 0)
	assert.Greater(t, matchmaking.CompareDuoPairsForTest(laterPair, earlierPair), 0)
	assert.Equal(t, 0, matchmaking.CompareDuoPairsForTest(earlierPair, earlierPair))

	if a.String() < b.String() {
		assert.Less(t, matchmaking.CompareDuoPairsForTest(
			matchmaking.DuoPairForTest{A: matchmaking.Player{UserID: a, CreatedAt: now}, B: matchmaking.Player{UserID: b, CreatedAt: now}},
			matchmaking.DuoPairForTest{A: matchmaking.Player{UserID: b, CreatedAt: now}, B: matchmaking.Player{UserID: a, CreatedAt: now}},
		), 0)
		assert.Greater(t, matchmaking.CompareDuoPairsForTest(
			matchmaking.DuoPairForTest{A: matchmaking.Player{UserID: b, CreatedAt: now}, B: matchmaking.Player{UserID: a, CreatedAt: now}},
			matchmaking.DuoPairForTest{A: matchmaking.Player{UserID: a, CreatedAt: now}, B: matchmaking.Player{UserID: b, CreatedAt: now}},
		), 0)
	}
}
