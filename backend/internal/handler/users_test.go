package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// UsersMeHandler — GET /users/me
// ============================================================

func TestUsersMeHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UsersMeHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUsersMeHandler_InvalidUserID(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me")
	// Set userID as a non-UUID string to trigger the parse error.
	c.Set("userID", "not-a-uuid")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UsersMeHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUsersMeHandler_StoreError(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me")
	test_util.WithUserIDString(c, userID)

	var ms *store.MockStore
	ms = &store.MockStore{
		GetUserByUserIDFn: func(_ context.Context, _ uuid.UUID) (model.User, error) {
			return model.User{}, errors.New("db error")
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UsersMeHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUsersMeHandler_Success(t *testing.T) {
	userID := uuid.New()
	username := "testuser"
	want := model.User{
		ID:          userID,
		DiscordName: &username,
	}

	c, w := test_util.NewGinContext(http.MethodGet, "/users/me")
	test_util.WithUserIDString(c, userID)

	var ms *store.MockStore
	ms = &store.MockStore{
		GetUserByUserIDFn: func(_ context.Context, id uuid.UUID) (model.User, error) {
			assert.Equal(t, userID, id)
			return want, nil
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UsersMeHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	got := test_util.DecodeJSON[model.User](t, w)
	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, want.DiscordName, got.DiscordName)
}

// ============================================================
// UsersMeGamesHandler — GET /users/me/games
// ============================================================

func TestUsersMeGamesHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/games")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UsersMeGamesHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUsersMeGamesHandler_InvalidUserID(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/games")
	c.Set("userID", "not-a-uuid")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UsersMeGamesHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUsersMeGamesHandler_StoreError(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/games")
	test_util.WithUserIDString(c, userID)

	var ms *store.MockStore
	ms = &store.MockStore{
		GetUserGamesForUserFn: func(_ context.Context, _ uuid.UUID) ([]model.UserGame, error) {
			return nil, errors.New("db error")
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UsersMeGamesHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUsersMeGamesHandler_Success(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	want := []model.UserGame{
		{GameID: gameID, UserID: userID},
	}

	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/games")
	test_util.WithUserIDString(c, userID)

	var ms *store.MockStore
	ms = &store.MockStore{
		GetUserGamesForUserFn: func(_ context.Context, id uuid.UUID) ([]model.UserGame, error) {
			assert.Equal(t, userID, id)
			return want, nil
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UsersMeGamesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	got := test_util.DecodeJSON[[]model.UserGame](t, w)
	require.Len(t, got, 1)
	assert.Equal(t, want[0].GameID, got[0].GameID)
}

func TestUsersMeGamesHandler_EmptyList(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/games")
	test_util.WithUserIDString(c, userID)

	var ms *store.MockStore
	ms = &store.MockStore{
		GetUserGamesForUserFn: func(_ context.Context, _ uuid.UUID) ([]model.UserGame, error) {
			return []model.UserGame{}, nil
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UsersMeGamesHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	got := test_util.DecodeJSON[[]model.UserGame](t, w)
	assert.Empty(t, got)
}

// ============================================================
// UsersMeEventsHandler — GET /users/me/events
// ============================================================

func TestUsersMeEventsHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/events?tz=UTC")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UsersMeEventsHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUsersMeEventsHandler_MissingTimezone(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/events")
	test_util.WithUserIDString(c, userID)

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UsersMeEventsHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUsersMeEventsHandler_InvalidFromDate(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/events?tz=UTC&from=2026/01/01")
	test_util.WithUserIDString(c, userID)

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UsersMeEventsHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUsersMeEventsHandler_InvalidTimezone(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/events?tz=Not_A_Timezone")
	test_util.WithUserIDString(c, userID)

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UsersMeEventsHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUsersMeEventsHandler_InvalidToDate(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/events?tz=UTC&to=2026/01/01")
	test_util.WithUserIDString(c, userID)

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UsersMeEventsHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUsersMeEventsHandler_InvalidDateRange(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/events?tz=UTC&from=2026-04-22&to=2026-04-21")
	test_util.WithUserIDString(c, userID)

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UsersMeEventsHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUsersMeEventsHandler_StoreValidationError(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/events?tz=UTC")
	test_util.WithUserIDString(c, userID)

	ms := &store.MockStore{
		GetEventsForUserFn: func(_ context.Context, _ uuid.UUID, _ bool, _ bool, _ *time.Time, _ *time.Time, _ string, _ string, _ string) ([]model.DashboardEvent, bool, string, error) {
			return nil, false, "", store.ErrInvalidCursor
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UsersMeEventsHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUsersMeEventsHandler_StoreServerError(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/events?tz=UTC")
	test_util.WithUserIDString(c, userID)

	ms := &store.MockStore{
		GetEventsForUserFn: func(_ context.Context, _ uuid.UUID, _ bool, _ bool, _ *time.Time, _ *time.Time, _ string, _ string, _ string) ([]model.DashboardEvent, bool, string, error) {
			return nil, false, "", errors.New("db down")
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UsersMeEventsHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUsersMeEventsHandler_Success(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.NewString()
	cursor := "opaque_cursor"
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/events?tz=UTC&hosting=true&past=true&from=2026-04-20&to=2026-04-21&game_id="+gameID+"&cursor="+cursor)
	test_util.WithUserIDString(c, userID)

	event := model.DashboardEvent{
		ID:               uuid.New(),
		Name:             "",
		GameName:         "Valorant",
		GameMode:         "5v5",
		Region:           "AMER",
		EventDate:        time.Now().UTC(),
		HostID:           userID,
		HostName:         "host",
		RegisteredCount:  10,
		RegistrationOpen: true,
	}

	ms := &store.MockStore{
		GetEventsForUserFn: func(_ context.Context, gotUserID uuid.UUID, hosting, past bool, from, to *time.Time, gotGameID, gotCursor, timezone string) ([]model.DashboardEvent, bool, string, error) {
			assert.Equal(t, userID, gotUserID)
			assert.True(t, hosting)
			assert.True(t, past)
			require.NotNil(t, from)
			require.NotNil(t, to)
			assert.Equal(t, "2026-04-20", from.Format("2006-01-02"))
			assert.Equal(t, "2026-04-21", to.Format("2006-01-02"))
			assert.Equal(t, gameID, gotGameID)
			assert.Equal(t, cursor, gotCursor)
			assert.Equal(t, "UTC", timezone)
			return []model.DashboardEvent{event}, true, "next_cursor", nil
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UsersMeEventsHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	type response struct {
		EventGroups []model.DashboardEvent `json:"event_groups"`
		NextCursor  string                 `json:"next_cursor"`
		HasMore     bool                   `json:"has_more"`
	}
	got := test_util.DecodeJSON[response](t, w)
	require.Len(t, got.EventGroups, 1)
	assert.Equal(t, event.ID, got.EventGroups[0].ID)
	assert.Equal(t, "next_cursor", got.NextCursor)
	assert.True(t, got.HasMore)
}

// ============================================================
// UpdateUsersMeHandler — PUT /users/me
// ============================================================

// validUpdateBody is a minimal valid request body for UpdateUsersMeHandler.
const validUpdateBody = `{"pronouns":"they/them","show_pronouns":true,"region":"EU","games":[]}`

func TestUpdateUsersMeHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPut, "/users/me")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UpdateUsersMeHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateUsersMeHandler_InvalidUserID(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPut, "/users/me")
	c.Set("userID", "not-a-uuid")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UpdateUsersMeHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateUsersMeHandler_InvalidBody(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodPut, "/users/me")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(`not json`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UpdateUsersMeHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUsersMeHandler_UpdateUserFails(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodPut, "/users/me")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(validUpdateBody))
	c.Request.Header.Set("Content-Type", "application/json")

	var ms *store.MockStore
	ms = &store.MockStore{
		WithTxFn: func(ctx context.Context, fn func(store.Store) error) error {
			return fn(ms) // execute fn with the mock itself as the tx store
		},
		UpdateUserFn: func(_ context.Context, _ uuid.UUID, _ *string, _ bool, _ *string) (model.User, error) {
			return model.User{}, errors.New("db error")
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UpdateUsersMeHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateUsersMeHandler_UpsertGameBadRequest(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	body := `{"pronouns":"they/them","show_pronouns":true,"region":"EU","games":[{"game_id":"` + gameID.String() + `"}]}`

	c, w := test_util.NewGinContext(http.MethodPut, "/users/me")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var ms *store.MockStore
	ms = &store.MockStore{
		WithTxFn: func(ctx context.Context, fn func(store.Store) error) error {
			return fn(ms)
		},
		UpdateUserFn: func(_ context.Context, _ uuid.UUID, _ *string, _ bool, _ *string) (model.User, error) {
			return model.User{ID: userID, UpdatedAt: time.Now()}, nil
		},
		UpsertGameForUserFn: func(_ context.Context, _ uuid.UUID, _ model.UserGame) (model.UserGame, error) {
			return model.UserGame{}, store.ErrInvalidGame
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UpdateUsersMeHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUsersMeHandler_UpsertGameServerError(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	body := `{"pronouns":"they/them","show_pronouns":true,"region":"EU","games":[{"game_id":"` + gameID.String() + `"}]}`

	c, w := test_util.NewGinContext(http.MethodPut, "/users/me")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var ms *store.MockStore
	ms = &store.MockStore{
		WithTxFn: func(ctx context.Context, fn func(store.Store) error) error {
			return fn(ms)
		},
		UpdateUserFn: func(_ context.Context, _ uuid.UUID, _ *string, _ bool, _ *string) (model.User, error) {
			return model.User{ID: userID, UpdatedAt: time.Now()}, nil
		},
		UpsertGameForUserFn: func(_ context.Context, _ uuid.UUID, _ model.UserGame) (model.UserGame, error) {
			return model.UserGame{}, errors.New("db error")
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UpdateUsersMeHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateUsersMeHandler_Success_NoGames(t *testing.T) {
	userID := uuid.New()
	username := "testuser"

	c, w := test_util.NewGinContext(http.MethodPut, "/users/me")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(validUpdateBody))
	c.Request.Header.Set("Content-Type", "application/json")

	var ms *store.MockStore
	ms = &store.MockStore{
		WithTxFn: func(ctx context.Context, fn func(store.Store) error) error {
			return fn(ms)
		},
		UpdateUserFn: func(_ context.Context, id uuid.UUID, _ *string, _ bool, _ *string) (model.User, error) {
			assert.Equal(t, userID, id)
			return model.User{ID: userID, DiscordName: &username}, nil
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UpdateUsersMeHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	type result struct {
		model.User
		Games []model.UserGame `json:"games"`
	}
	got := test_util.DecodeJSON[result](t, w)
	assert.Equal(t, userID, got.ID)
	assert.Empty(t, got.Games)
}

func TestUpdateUsersMeHandler_Success_WithGames(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	username := "testuser"
	body := `{"pronouns":"they/them","show_pronouns":true,"region":"EU","games":[{"game_id":"` + gameID.String() + `"}]}`

	c, w := test_util.NewGinContext(http.MethodPut, "/users/me")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var ms *store.MockStore
	ms = &store.MockStore{
		WithTxFn: func(ctx context.Context, fn func(store.Store) error) error {
			return fn(ms)
		},
		UpdateUserFn: func(_ context.Context, _ uuid.UUID, _ *string, _ bool, _ *string) (model.User, error) {
			return model.User{ID: userID, DiscordName: &username}, nil
		},
		UpsertGameForUserFn: func(_ context.Context, _ uuid.UUID, ug model.UserGame) (model.UserGame, error) {
			return model.UserGame{GameID: ug.GameID, UserID: userID}, nil
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.UpdateUsersMeHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	type result struct {
		model.User
		Games []model.UserGame `json:"games"`
	}
	got := test_util.DecodeJSON[result](t, w)
	assert.Equal(t, userID, got.ID)
	require.Len(t, got.Games, 1)
	assert.Equal(t, gameID, got.Games[0].GameID)
}

func TestUpdateUsersMeHandler_UsesTxStoreInsideWithTx(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	username := "tx-user"
	body := `{"pronouns":"they/them","show_pronouns":true,"region":"EU","games":[{"game_id":"` + gameID.String() + `"}]}`

	c, w := test_util.NewGinContext(http.MethodPut, "/users/me")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	txStore := &store.MockStore{
		UpdateUserFn: func(_ context.Context, id uuid.UUID, _ *string, _ bool, _ *string) (model.User, error) {
			assert.Equal(t, userID, id)
			return model.User{ID: userID, DiscordName: &username}, nil
		},
		UpsertGameForUserFn: func(_ context.Context, id uuid.UUID, ug model.UserGame) (model.UserGame, error) {
			assert.Equal(t, userID, id)
			return model.UserGame{GameID: ug.GameID, UserID: id}, nil
		},
	}

	baseUpdateCalled := false
	baseUpsertCalled := false
	baseStore := &store.MockStore{
		WithTxFn: func(_ context.Context, fn func(store.Store) error) error {
			return fn(txStore)
		},
		UpdateUserFn: func(_ context.Context, _ uuid.UUID, _ *string, _ bool, _ *string) (model.User, error) {
			baseUpdateCalled = true
			return model.User{}, errors.New("base store should not be used in tx callback")
		},
		UpsertGameForUserFn: func(_ context.Context, _ uuid.UUID, _ model.UserGame) (model.UserGame, error) {
			baseUpsertCalled = true
			return model.UserGame{}, errors.New("base store should not be used in tx callback")
		},
	}

	h := newTestHandler(t, baseStore, nil, "")
	h.UpdateUsersMeHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, baseUpdateCalled, "UpdateUser must run on tx store")
	assert.False(t, baseUpsertCalled, "UpsertGameForUser must run on tx store")
}

// ============================================================
// UpsertUsersMeGameHandler — PUT /users/me/games/:gameId
// ============================================================

func TestUpsertUsersMeGameHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPut, "/users/me/games/x")
	c.Params = gin.Params{{Key: "gameId", Value: uuid.NewString()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"in_game_name":"abc","current_rank":"` + uuid.NewString() + `","peak_rank":"` + uuid.NewString() + `","show_rank":true}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UpsertUsersMeGameHandler(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpsertUsersMeGameHandler_Validation(t *testing.T) {
	t.Run("invalid game id", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/users/me/games/bad")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "gameId", Value: "bad"}}
		c.Request.Body = io.NopCloser(strings.NewReader(`{"in_game_name":"abc","current_rank":"` + uuid.NewString() + `","peak_rank":"` + uuid.NewString() + `","show_rank":true}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertUsersMeGameHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad json", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/users/me/games/x")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "gameId", Value: uuid.NewString()}}
		c.Request.Body = io.NopCloser(strings.NewReader(`{`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertUsersMeGameHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid current rank", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/users/me/games/x")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "gameId", Value: uuid.NewString()}}
		c.Request.Body = io.NopCloser(strings.NewReader(`{"in_game_name":"abc","current_rank":"bad","peak_rank":"` + uuid.NewString() + `","show_rank":true}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertUsersMeGameHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid peak rank", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/users/me/games/x")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "gameId", Value: uuid.NewString()}}
		c.Request.Body = io.NopCloser(strings.NewReader(`{"in_game_name":"abc","current_rank":"` + uuid.NewString() + `","peak_rank":"bad","show_rank":true}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertUsersMeGameHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing in game name", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/users/me/games/x")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "gameId", Value: uuid.NewString()}}
		c.Request.Body = io.NopCloser(strings.NewReader(`{"in_game_name":"  ","current_rank":"` + uuid.NewString() + `","peak_rank":"` + uuid.NewString() + `","show_rank":true}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertUsersMeGameHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpsertUsersMeGameHandler_StoreErrors(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	currentRankID := uuid.New()
	peakRankID := uuid.New()

	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"invalid payload", store.ErrInvalidCurrentRank, http.StatusBadRequest},
		{"server error", errors.New("db exploded"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test_util.NewGinContext(http.MethodPut, "/users/me/games/x")
			test_util.WithUserIDString(c, userID)
			c.Params = gin.Params{{Key: "gameId", Value: gameID.String()}}
			c.Request.Body = io.NopCloser(strings.NewReader(`{"in_game_name":"  PlayerOne  ","current_rank":"` + currentRankID.String() + `","peak_rank":"` + peakRankID.String() + `","show_rank":true}`))
			c.Request.Header.Set("Content-Type", "application/json")

			h := newTestHandler(t, &store.MockStore{
				UpsertGameForUserFn: func(_ context.Context, inUser uuid.UUID, ug model.UserGame) (model.UserGame, error) {
					assert.Equal(t, userID, inUser)
					assert.Equal(t, gameID, ug.GameID)
					require.NotNil(t, ug.CurrentRank)
					require.NotNil(t, ug.PeakRank)
					assert.Equal(t, currentRankID, *ug.CurrentRank)
					assert.Equal(t, peakRankID, *ug.PeakRank)
					require.NotNil(t, ug.InGameName)
					assert.Equal(t, "PlayerOne", *ug.InGameName)
					assert.True(t, ug.ShowRank)
					return model.UserGame{}, tc.err
				},
			}, nil, "")
			h.UpsertUsersMeGameHandler(c)
			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUpsertUsersMeGameHandler_Success(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.New()
	currentRankID := uuid.New()
	peakRankID := uuid.New()

	c, _ := test_util.NewGinContext(http.MethodPut, "/users/me/games/x")
	test_util.WithUserIDString(c, userID)
	c.Params = gin.Params{{Key: "gameId", Value: gameID.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"in_game_name":"  PlayerOne  ","current_rank":"` + currentRankID.String() + `","peak_rank":"` + peakRankID.String() + `","show_rank":false}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := newTestHandler(t, &store.MockStore{
		UpsertGameForUserFn: func(_ context.Context, inUser uuid.UUID, ug model.UserGame) (model.UserGame, error) {
			assert.Equal(t, userID, inUser)
			assert.Equal(t, gameID, ug.GameID)
			require.NotNil(t, ug.InGameName)
			assert.Equal(t, "PlayerOne", *ug.InGameName)
			require.NotNil(t, ug.CurrentRank)
			require.NotNil(t, ug.PeakRank)
			assert.Equal(t, currentRankID, *ug.CurrentRank)
			assert.Equal(t, peakRankID, *ug.PeakRank)
			assert.False(t, ug.ShowRank)
			return ug, nil
		},
	}, nil, "")
	h.UpsertUsersMeGameHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}
