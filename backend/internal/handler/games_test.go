package handler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/handler"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Handler tests — GetSystemGamesHandler
// ============================================================

func TestGetSystemGamesHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/games")

	h := newTestHandler(t, &store.MockStore{}, nil)
	h.GetSystemGamesHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetSystemGamesHandler_StoreError(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/games")
	test_util.WithUserID(c, uuid.New())

	ms := &store.MockStore{
		GetSystemGamesFn: func(_ context.Context) ([]model.Game, error) {
			return nil, errors.New("db exploded")
		},
	}
	h := newTestHandler(t, ms, nil)
	h.GetSystemGamesHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetSystemGamesHandler_Success(t *testing.T) {
	want := []model.Game{
		{ID: uuid.New(), Name: "Valorant"},
		{ID: uuid.New(), Name: "Apex Legends"},
	}

	c, w := test_util.NewGinContext(http.MethodGet, "/games")
	test_util.WithUserID(c, uuid.New())

	ms := &store.MockStore{
		GetSystemGamesFn: func(_ context.Context) ([]model.Game, error) {
			return want, nil
		},
	}
	h := newTestHandler(t, ms, nil)
	h.GetSystemGamesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	got := test_util.DecodeJSON[[]model.Game](t, w)
	assert.Len(t, got, len(want))
	assert.Equal(t, want[0].Name, got[0].Name)
	assert.Equal(t, want[1].Name, got[1].Name)
}

func TestGetSystemGamesHandler_EmptyList(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/games")
	test_util.WithUserID(c, uuid.New())

	ms := &store.MockStore{
		GetSystemGamesFn: func(_ context.Context) ([]model.Game, error) {
			return []model.Game{}, nil
		},
	}
	h := newTestHandler(t, ms, nil)
	h.GetSystemGamesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	got := test_util.DecodeJSON[[]model.Game](t, w)
	assert.Empty(t, got)
}

// ============================================================
// Handler tests — GetGameRanksByGame
// ============================================================

func TestGetGameRanksByGame_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/games/abc/ranks")

	h := newTestHandler(t, &store.MockStore{}, nil)
	h.GetGameRanksByGame(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetGameRanksByGame_InvalidGameID(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/games/not-a-uuid/ranks")
	test_util.WithUserID(c, uuid.New())
	c.Params = gin.Params{{Key: "gameId", Value: "not-a-uuid"}}

	h := newTestHandler(t, &store.MockStore{}, nil)
	h.GetGameRanksByGame(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetGameRanksByGame_StoreError(t *testing.T) {
	gameID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/games/"+gameID.String()+"/ranks")
	test_util.WithUserID(c, uuid.New())
	c.Params = gin.Params{{Key: "gameId", Value: gameID.String()}}

	ms := &store.MockStore{
		GetGameRanksFn: func(_ context.Context, _ *uuid.UUID) ([]model.GameRank, error) {
			return nil, errors.New("db exploded")
		},
	}
	h := newTestHandler(t, ms, nil)
	h.GetGameRanksByGame(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetGameRanksByGame_Success(t *testing.T) {
	gameID := uuid.New()
	want := []model.GameRank{
		{ID: uuid.New(), GameID: &gameID, Name: "Bronze", Order: 1},
		{ID: uuid.New(), GameID: &gameID, Name: "Silver", Order: 2},
		{ID: uuid.New(), GameID: &gameID, Name: "Gold", Order: 3},
	}

	c, w := test_util.NewGinContext(http.MethodGet, "/games/"+gameID.String()+"/ranks")
	test_util.WithUserID(c, uuid.New())
	c.Params = gin.Params{{Key: "gameId", Value: gameID.String()}}

	ms := &store.MockStore{
		GetGameRanksFn: func(_ context.Context, id *uuid.UUID) ([]model.GameRank, error) {
			assert.Equal(t, gameID, *id)
			return want, nil
		},
	}
	h := newTestHandler(t, ms, nil)
	h.GetGameRanksByGame(c)

	assert.Equal(t, http.StatusOK, w.Code)
	got := test_util.DecodeJSON[[]model.GameRank](t, w)
	assert.Len(t, got, len(want))
	for i, r := range got {
		assert.Equal(t, want[i].Name, r.Name)
		assert.Equal(t, want[i].Order, r.Order)
	}
}

// ============================================================
// Integration tests — business logic against a real DB tx
// ============================================================

func TestGetSystemGames_Integration(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		games, err := handler.GetSystemGames(s, context.Background())
		require.NoError(t, err)
		assert.NotNil(t, games)
	})
}

func TestGetGameRanksForGame_Integration_UnknownGame(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		ranks, err := handler.GetGameRanksForGame(uuid.New(), s, context.Background())
		require.NoError(t, err)
		assert.Empty(t, ranks)
	})
}

func TestGetGameRanksForGame_Integration_KnownGame(t *testing.T) {
	test_util.WithTestTx(t, func(_ *db.Queries, s *store.PostgresStore) {
		games, err := handler.GetSystemGames(s, context.Background())
		require.NoError(t, err)

		if len(games) == 0 {
			t.Skip("no system games seeded — skipping rank integration test")
		}

		ranks, err := handler.GetGameRanksForGame(games[0].ID, s, context.Background())
		require.NoError(t, err)
		assert.NotNil(t, ranks)
	})
}
