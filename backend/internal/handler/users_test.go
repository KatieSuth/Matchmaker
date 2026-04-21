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

func TestUsersMeEventsHandler_Success(t *testing.T) {
	userID := uuid.New()
	gameID := uuid.NewString()
	cursor := "opaque_cursor"
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/events?tz=UTC&hosting=true&past=true&from=2026-04-20&to=2026-04-21&game_id="+gameID+"&cursor="+cursor)
	test_util.WithUserIDString(c, userID)

	event := model.DashboardEvent{
		ID:               uuid.New(),
		GameName:         "Valorant",
		GameMode:         "5v5",
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
		UpsertGameForUserFn: func(_ context.Context, _ uuid.UUID, _ model.UserGame) (model.UserGame, int, error) {
			return model.UserGame{}, http.StatusBadRequest, errors.New("invalid game data")
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
		UpsertGameForUserFn: func(_ context.Context, _ uuid.UUID, _ model.UserGame) (model.UserGame, int, error) {
			return model.UserGame{}, http.StatusInternalServerError, errors.New("db error")
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
		UpsertGameForUserFn: func(_ context.Context, _ uuid.UUID, ug model.UserGame) (model.UserGame, int, error) {
			return model.UserGame{GameID: ug.GameID, UserID: userID}, http.StatusOK, nil
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
