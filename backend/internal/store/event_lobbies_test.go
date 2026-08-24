package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gameRanksForGame(t *testing.T, ctx context.Context, tx db.DBTX, gameID uuid.UUID) []uuid.UUID {
	t.Helper()
	rows, err := tx.Query(ctx, `SELECT id FROM game_ranks WHERE game_id = $1 ORDER BY "order" ASC`, gameID)
	require.NoError(t, err)
	defer rows.Close()

	var rankIDs []uuid.UUID
	for rows.Next() {
		var rankID uuid.UUID
		require.NoError(t, rows.Scan(&rankID))
		rankIDs = append(rankIDs, rankID)
	}
	require.NoError(t, rows.Err())
	require.GreaterOrEqual(t, len(rankIDs), 2, "need at least 2 ranks to test team ordering")
	return rankIDs
}

func registerPlayerForEventWithRank(
	t *testing.T,
	ctx context.Context,
	tx db.DBTX,
	s *store.PostgresStore,
	eventID, userID, gameID, rankID uuid.UUID,
	canSubstitute, canLobbyHost bool,
) {
	t.Helper()
	inGameName := "ranked-player"
	_, err := s.UpsertGameForUser(ctx, userID, model.UserGame{
		GameID:      gameID,
		InGameName:  &inGameName,
		CurrentRank: &rankID,
		PeakRank:    &rankID,
		ShowRank:    true,
	})
	require.NoError(t, err)
	registerUserForEventWithFlags(t, ctx, tx, eventID, userID, canSubstitute, canLobbyHost)
}

func assertTeamPlayersSortedByRankDesc(t *testing.T, players []model.LobbyPlayer) {
	t.Helper()
	for i := 1; i < len(players); i++ {
		prev := players[i-1]
		curr := players[i]
		if prev.CurrentRankOrder == curr.CurrentRankOrder {
			assert.LessOrEqual(t, prev.DiscordName, curr.DiscordName)
			continue
		}
		assert.Greater(t, prev.CurrentRankOrder, curr.CurrentRankOrder)
	}
}

func TestGetEventGroupDetail_TeamPlayersSortedByRankDesc(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	ranks := gameRanksForGame(t, ctx, tx, games[0].ID)
	lowRank := ranks[0]
	highRank := ranks[len(ranks)-1]

	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	players := make([]uuid.UUID, 0, 4)
	for i, rankID := range []uuid.UUID{lowRank, highRank, ranks[len(ranks)/2], lowRank} {
		u := createTestUser(t, ctx, s)
		players = append(players, u.ID)
		registerPlayerForEventWithRank(t, ctx, tx, s, eventID, u.ID, games[0].ID, rankID, false, i == 0)
	}
	registerPlayerForEventWithRank(t, ctx, tx, s, eventID, host.ID, games[0].ID, ranks[1], false, false)

	_, err = s.CreateTeamsForGroup(ctx, groupID, host.ID, defaultMatchmakingSettings())
	require.NoError(t, err)

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	require.Len(t, detail.Events, 1)
	require.Len(t, detail.Events[0].Lobbies, 1)

	for _, team := range detail.Events[0].Lobbies[0].Teams {
		assertTeamPlayersSortedByRankDesc(t, team.Players)
	}
}
