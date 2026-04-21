package handler_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/jackc/pgx/v5"
)

const validCreateEventBody = `{
  "game_mode_id": "%s",
  "region": "NA",
  "start_time": "%s",
  "sub_min": 0,
  "games_to_run": 3,
  "registration_open": true
}`

func startTimeHoursFromNow(h int) string {
	return time.Now().UTC().Add(time.Duration(h) * time.Hour).Truncate(time.Second).Format(time.RFC3339)
}

func TestCreateEventHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateEventHandler_InvalidUserID(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	c.Set("userID", "not-a-uuid")
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateEventHandler_BadJSON(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader("not-json"))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateEventHandler_ValidationFailure(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	body := `{
	  "game_mode_id": "` + gameModeID.String() + `",
	  "region": "",
	  "start_time": "2026-04-20T18:00:00Z",
	  "sub_min": 0,
	  "games_to_run": 0,
	  "registration_open": true
	}`

	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateEventHandler_StoreError(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	body := sprintf(validCreateEventBody, gameModeID.String(), startTimeHoursFromNow(48))

	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ms := &store.MockStore{
		CreateEventGroupWithEventsFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int32, _ bool, _ string, _ time.Time, _ int32) (uuid.UUID, error) {
			return uuid.Nil, errors.New("db exploded")
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateEventHandler_GameModeNotFound(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	body := sprintf(validCreateEventBody, gameModeID.String(), startTimeHoursFromNow(48))

	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ms := &store.MockStore{
		CreateEventGroupWithEventsFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int32, _ bool, _ string, _ time.Time, _ int32) (uuid.UUID, error) {
			return uuid.Nil, pgx.ErrNoRows
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateEventHandler_SubMinNegative(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	body := `{
	  "game_mode_id": "` + gameModeID.String() + `",
	  "region": "NA",
	  "start_time": "` + startTimeHoursFromNow(48) + `",
	  "sub_min": -1,
	  "games_to_run": 3,
	  "registration_open": true
	}`

	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateEventHandler_StartTimeInPast(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	body := `{
	  "game_mode_id": "` + gameModeID.String() + `",
	  "region": "NA",
	  "start_time": "2000-01-01T00:00:00Z",
	  "sub_min": 0,
	  "games_to_run": 3,
	  "registration_open": true
	}`

	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateEventHandler_Success(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	groupID := uuid.New()
	startAt := startTimeHoursFromNow(48)
	body := sprintf(validCreateEventBody, gameModeID.String(), startAt)

	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ms := &store.MockStore{
		CreateEventGroupWithEventsFn: func(_ context.Context, inUserID, inGameModeID uuid.UUID, subMin int32, registrationOpen bool, region string, startTime time.Time, gamesToRun int32) (uuid.UUID, error) {
			assert.Equal(t, userID, inUserID)
			assert.Equal(t, gameModeID, inGameModeID)
			assert.Equal(t, int32(0), subMin)
			assert.True(t, registrationOpen)
			assert.Equal(t, "NA", region)
			assert.Equal(t, int32(3), gamesToRun)
			want, err := time.Parse(time.RFC3339, startAt)
			require.NoError(t, err)
			assert.Equal(t, want.UTC().Format(time.RFC3339), startTime.UTC().Format(time.RFC3339))
			return groupID, nil
		},
	}
	h := newTestHandler(t, ms, nil, "")
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	type response struct {
		GroupID string `json:"group_id"`
	}
	got := test_util.DecodeJSON[response](t, w)
	require.Equal(t, groupID.String(), got.GroupID)
}

func sprintf(format string, args ...interface{}) string {
	return strings.TrimSpace(fmt.Sprintf(format, args...))
}
