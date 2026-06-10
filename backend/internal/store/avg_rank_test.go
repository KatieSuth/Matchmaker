package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gameRankExtremes(t *testing.T, ctx context.Context, s *store.PostgresStore, gameID uuid.UUID) (low, high model.GameRank) {
	t.Helper()
	ranks, err := s.GetGameRanks(ctx, &gameID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(ranks), 2, "need at least 2 ranks")
	for _, r := range ranks {
		if low.ID == uuid.Nil || r.Order < low.Order {
			low = r
		}
		if high.ID == uuid.Nil || r.Order > high.Order {
			high = r
		}
	}
	return low, high
}

func expectedAvgRankID(t *testing.T, ranks []model.GameRank, lowOrder, highOrder int32) uuid.UUID {
	t.Helper()
	floored := int32(matchmaking.FlooredAverageRankOrder(int(lowOrder), int(highOrder)))
	for _, r := range ranks {
		if r.Order == floored {
			return r.ID
		}
	}
	t.Fatalf("no rank at floored order %d", floored)
	return uuid.Nil
}

func TestResolveAvgRankIDForTest_Success(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()

	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)

	low, high := gameRankExtremes(t, ctx, s, games[0].ID)
	modelRanks, err := s.GetGameRanks(ctx, &games[0].ID)
	require.NoError(t, err)
	wantID := expectedAvgRankID(t, modelRanks, low.Order, high.Order)
	wantOrder := int32(matchmaking.FlooredAverageRankOrder(int(low.Order), int(high.Order)))

	gotID, gotOrder, err := store.ResolveAvgRankIDForTest(s, ctx, games[0].ID, low.Order, high.Order)
	require.NoError(t, err)
	assert.Equal(t, wantID, gotID)
	assert.Equal(t, wantOrder, gotOrder)
}

func TestResolveAvgRankIDForTest_NoRankAtOrder(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()

	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)

	_, _, err = store.ResolveAvgRankIDForTest(s, ctx, games[0].ID, 99998, 99999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no rank for game")
	assert.Contains(t, err.Error(), games[0].ID.String())
}

func TestResolveAvgRankIDForTest_LookupError(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()

	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)
	low, high := gameRankExtremes(t, ctx, s, games[0].ID)

	faulty := &faultInjectTx{
		DBTX:               tx,
		failOnQueryRowCall: 1,
		injectedErr:        injectedErr,
	}
	fs := store.NewPostgresStoreFromDBTXForTest(faulty)

	_, _, err = store.ResolveAvgRankIDForTest(fs, ctx, games[0].ID, low.Order, high.Order)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lookup avg rank for game")
	assert.ErrorIs(t, err, injectedErr)
}

func TestEnsureAvgRanksForMatchmakingForTest_SkipsWhenPresent(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()

	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)
	low, high := gameRankExtremes(t, ctx, s, games[0].ID)
	modelRanks, err := s.GetGameRanks(ctx, &games[0].ID)
	require.NoError(t, err)
	avgID := expectedAvgRankID(t, modelRanks, low.Order, high.Order)
	avgOrder := int32(matchmaking.FlooredAverageRankOrder(int(low.Order), int(high.Order)))

	rows := []db.GetMatchmakingRegistrationsForEventRow{{
		UserID:           uuid.New(),
		GameID:           games[0].ID,
		AvgRank:          &avgID,
		AvgRankOrder:     &avgOrder,
		CurrentRankOrder: low.Order,
		PeakRankOrder:    high.Order,
	}}

	faulty := &faultInjectTx{
		DBTX:           tx,
		failOnExecCall: 1,
		injectedErr:    injectedErr,
	}
	fs := store.NewPostgresStoreFromDBTXForTest(faulty)

	got, err := store.EnsureAvgRanksForMatchmakingForTest(fs, ctx, rows)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, avgID, *got[0].AvgRank)
	assert.Equal(t, avgOrder, *got[0].AvgRankOrder)
}

func TestEnsureAvgRanksForMatchmakingForTest_BackfillsMissing(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()

	host := createTestUser(t, ctx, s)
	participant := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	_, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, time.Now().UTC().Add(24*time.Hour))

	low, high := gameRankExtremes(t, ctx, s, games[0].ID)
	modelRanks, err := s.GetGameRanks(ctx, &games[0].ID)
	require.NoError(t, err)
	wantAvgID := expectedAvgRankID(t, modelRanks, low.Order, high.Order)
	wantOrder := int32(matchmaking.FlooredAverageRankOrder(int(low.Order), int(high.Order)))

	inGameName := "avg-backfill"
	_, err = s.UpsertGameForUser(ctx, participant.ID, model.UserGame{
		GameID:      games[0].ID,
		CurrentRank: &low.ID,
		PeakRank:    &high.ID,
		InGameName:  &inGameName,
		ShowRank:    true,
	})
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `
		UPDATE user_games
		SET avg_rank = NULL
		WHERE user_id = $1 AND game_id = $2
	`, participant.ID, games[0].ID)
	require.NoError(t, err)

	registerUserForEvent(t, ctx, tx, eventID, participant.ID)

	q := db.New(tx)
	regRows, err := q.GetMatchmakingRegistrationsForEvent(ctx, eventID)
	require.NoError(t, err)
	require.Len(t, regRows, 1)
	assert.Nil(t, regRows[0].AvgRank)
	assert.Nil(t, regRows[0].AvgRankOrder)

	got, err := store.EnsureAvgRanksForMatchmakingForTest(s, ctx, regRows)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.NotNil(t, got[0].AvgRank)
	require.NotNil(t, got[0].AvgRankOrder)
	assert.Equal(t, wantAvgID, *got[0].AvgRank)
	assert.Equal(t, wantOrder, *got[0].AvgRankOrder)

	var storedAvg uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT avg_rank FROM user_games WHERE user_id = $1 AND game_id = $2
	`, participant.ID, games[0].ID).Scan(&storedAvg)
	require.NoError(t, err)
	assert.Equal(t, wantAvgID, storedAvg)
}

func TestEnsureAvgRanksForMatchmakingForTest_ResolveError(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()

	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)

	rows := []db.GetMatchmakingRegistrationsForEventRow{{
		UserID:           uuid.New(),
		GameID:           games[0].ID,
		CurrentRankOrder: 99998,
		PeakRankOrder:    99999,
	}}

	_, err = store.EnsureAvgRanksForMatchmakingForTest(s, ctx, rows)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no rank for game")
}

func TestEnsureAvgRanksForMatchmakingForTest_PersistError(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()

	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)
	low, high := gameRankExtremes(t, ctx, s, games[0].ID)

	rows := []db.GetMatchmakingRegistrationsForEventRow{{
		UserID:           uuid.New(),
		GameID:           games[0].ID,
		CurrentRankOrder: low.Order,
		PeakRankOrder:    high.Order,
	}}

	// resolve succeeds on real tx; only Exec (UpdateUserGameAvgRank) fails.
	faulty := &faultInjectTx{
		DBTX:           tx,
		failOnExecCall: 1,
		injectedErr:    injectedErr,
	}
	fs := store.NewPostgresStoreFromDBTXForTest(faulty)

	_, err = store.EnsureAvgRanksForMatchmakingForTest(fs, ctx, rows)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update avg rank for user")
	assert.ErrorIs(t, err, injectedErr)
}
