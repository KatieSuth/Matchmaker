package matchmaking_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanEvent_SingleLobby(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 4)
	for i := 0; i < 4; i++ {
		players = append(players, matchmaking.Player{
			UserID:    uuid.New(),
			AvgRank:   float64(i + 1),
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		})
	}

	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    0,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}

	plan, err := matchmaking.PlanEvent(players, cfg, settings)
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)
	assert.Len(t, plan.Lobbies[0].Roster, 4)
}

func TestPlanEvent_RankedMode(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 4)
	for i := 0; i < 4; i++ {
		players = append(players, matchmaking.Player{
			UserID:    uuid.New(),
			AvgRank:   float64(i + 1),
			CreatedAt: now.Add(time.Duration(i) * time.Minute),
		})
	}

	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    0,
		SortLogic: "ranked",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}

	plan, err := matchmaking.PlanEvent(players, cfg, settings)
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)
	assert.NotNil(t, plan.Lobbies[0].HostID)
}

func TestPlanEvent_MultiLobbyWithMandatorySubs(t *testing.T) {
	now := time.Now()
	players := make([]matchmaking.Player, 0, 10)
	for i := 0; i < 8; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       float64(i + 1),
			CanSubstitute: false,
			CreatedAt:     now.Add(time.Duration(i) * time.Minute),
		})
	}
	for i := 0; i < 2; i++ {
		players = append(players, matchmaking.Player{
			UserID:        uuid.New(),
			AvgRank:       float64(20 + i),
			CanSubstitute: true,
			CreatedAt:     now.Add(time.Duration(8+i) * time.Minute),
		})
	}

	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    1,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}

	plan, err := matchmaking.PlanEvent(players, cfg, settings)
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 2)
	assert.Len(t, plan.Lobbies[0].Subs, 1)
	assert.Len(t, plan.Lobbies[1].Subs, 1)
}

func TestPlanEvent_InsufficientPlayers(t *testing.T) {
	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  5,
		SubMin:    0,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (5v5)",
		Slots:     10,
	}
	players := []matchmaking.Player{{UserID: uuid.New(), AvgRank: 10, CreatedAt: time.Now()}}

	_, err := matchmaking.PlanEvent(players, cfg, matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	})
	require.Error(t, err)
	var valErr *matchmaking.ValidationError
	require.ErrorAs(t, err, &valErr)
}

func TestPlanEvent_PreservesMutualDuoOnSameTeamWhenSnakeAllows(t *testing.T) {
	now := time.Now()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()

	players := []matchmaking.Player{
		{UserID: a, DiscordName: "A", AvgRank: 20, CreatedAt: now},
		{UserID: b, DiscordName: "B", DuoRequest: duoRequest("C"), AvgRank: 18, CreatedAt: now.Add(time.Minute)},
		{UserID: c, DiscordName: "C", DuoRequest: duoRequest("B"), AvgRank: 17, CreatedAt: now.Add(2 * time.Minute)},
		{UserID: d, DiscordName: "D", AvgRank: 10, CreatedAt: now.Add(3 * time.Minute)},
	}

	cfg := matchmaking.Config{
		EventID:   uuid.New(),
		TeamSize:  2,
		SubMin:    0,
		SortLogic: "balanced",
		TierCount: 25,
		GameLabel: "Game 1 (2v2)",
		Slots:     4,
	}
	settings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}

	plan, err := matchmaking.PlanEvent(players, cfg, settings)
	require.NoError(t, err)
	require.Len(t, plan.Lobbies, 1)

	teamFor := func(id uuid.UUID) int {
		for _, p := range plan.Lobbies[0].Roster {
			if p.UserID != id || p.TeamNumber == nil {
				continue
			}
			return *p.TeamNumber
		}
		return 0
	}
	assert.Equal(t, teamFor(b), teamFor(c))
}
