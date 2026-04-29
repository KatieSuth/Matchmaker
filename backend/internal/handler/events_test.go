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

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validCreateEventBody = `{
  "game_mode_id": "%s",
  "region": "AMER",
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
	  "region": "AMER",
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
	  "region": "AMER",
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
			assert.Equal(t, "AMER", region)
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

func TestGetEventGroupHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/events/x")
	c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.GetEventGroupHandler(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetEventGroupHandler_InvalidGroupID(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/events/bad")
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: "not-uuid"}}
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.GetEventGroupHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetEventGroupHandler_NotFound(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/events/x")
	gid := uuid.New()
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
	h := newTestHandler(t, &store.MockStore{
		GetEventGroupDetailFn: func(_ context.Context, inGroup, _ uuid.UUID) (model.EventGroupDetail, error) {
			assert.Equal(t, gid, inGroup)
			return model.EventGroupDetail{}, pgx.ErrNoRows
		},
	}, nil, "")
	h.GetEventGroupHandler(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetEventGroupHandler_StoreError(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/events/x")
	gid := uuid.New()
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
	h := newTestHandler(t, &store.MockStore{
		GetEventGroupDetailFn: func(_ context.Context, _, _ uuid.UUID) (model.EventGroupDetail, error) {
			return model.EventGroupDetail{}, errors.New("db error")
		},
	}, nil, "")
	h.GetEventGroupHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetEventGroupHandler_Success(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/events/x")
	gid := uuid.New()
	viewer := uuid.New()
	test_util.WithUserIDString(c, viewer)
	c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
	detail := model.EventGroupDetail{ID: gid, Region: "AMER"}
	h := newTestHandler(t, &store.MockStore{
		GetEventGroupDetailFn: func(_ context.Context, inGroup, inViewer uuid.UUID) (model.EventGroupDetail, error) {
			assert.Equal(t, gid, inGroup)
			assert.Equal(t, viewer, inViewer)
			return detail, nil
		},
	}, nil, "")
	h.GetEventGroupHandler(c)
	assert.Equal(t, http.StatusOK, w.Code)
	got := test_util.DecodeJSON[model.EventGroupDetail](t, w)
	assert.Equal(t, gid, got.ID)
	assert.Equal(t, "AMER", got.Region)
}

func TestUpdateEventGroupSettingsHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPatch, "/events/x")
	c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UpdateEventGroupSettingsHandler(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateEventGroupSettingsHandler_InvalidGroupID(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPatch, "/events/x")
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: "bad"}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"region":"AMER","sub_min":0}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UpdateEventGroupSettingsHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateEventGroupSettingsHandler_BadJSON(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPatch, "/events/x")
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UpdateEventGroupSettingsHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateEventGroupSettingsHandler_StoreErrors(t *testing.T) {
	gid := uuid.New()
	uid := uuid.New()
	body := `{"region":"AMER","sub_min":1}`

	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"forbidden", store.ErrForbidden, http.StatusForbidden},
		{"invalid", store.ErrInvalidSubMin, http.StatusBadRequest},
		{"not found", pgx.ErrNoRows, http.StatusNotFound},
		{"other", errors.New("fail"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test_util.NewGinContext(http.MethodPatch, "/events/x")
			test_util.WithUserIDString(c, uid)
			c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
			c.Request.Body = io.NopCloser(strings.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			h := newTestHandler(t, &store.MockStore{
				UpdateEventGroupSettingsFn: func(_ context.Context, inG, inO uuid.UUID, _ string, _ int32) error {
					assert.Equal(t, gid, inG)
					assert.Equal(t, uid, inO)
					return tc.err
				},
			}, nil, "")
			h.UpdateEventGroupSettingsHandler(c)
			assert.Equal(t, tc.status, w.Code, tc.name)
		})
	}
}

func TestUpdateEventGroupSettingsHandler_Success(t *testing.T) {
	gid := uuid.New()
	uid := uuid.New()
	c, _ := test_util.NewGinContext(http.MethodPatch, "/events/x")
	test_util.WithUserIDString(c, uid)
	c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"region":"EU","sub_min":2}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{
		UpdateEventGroupSettingsFn: func(_ context.Context, inG, inO uuid.UUID, region string, sub int32) error {
			assert.Equal(t, "EU", region)
			assert.Equal(t, int32(2), sub)
			return nil
		},
	}, nil, "")
	h.UpdateEventGroupSettingsHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestDeleteEventGroupHandler_StoreErrors(t *testing.T) {
	gid := uuid.New()
	uid := uuid.New()
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"forbidden", store.ErrForbidden, http.StatusForbidden},
		{"not found", pgx.ErrNoRows, http.StatusNotFound},
		{"other", errors.New("fail"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test_util.NewGinContext(http.MethodDelete, "/events/x")
			test_util.WithUserIDString(c, uid)
			c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
			h := newTestHandler(t, &store.MockStore{
				DeleteEventGroupFn: func(_ context.Context, inG, inO uuid.UUID) error {
					return tc.err
				},
			}, nil, "")
			h.DeleteEventGroupHandler(c)
			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestDeleteEventGroupHandler_Success(t *testing.T) {
	gid := uuid.New()
	uid := uuid.New()
	c, _ := test_util.NewGinContext(http.MethodDelete, "/events/x")
	test_util.WithUserIDString(c, uid)
	c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
	h := newTestHandler(t, &store.MockStore{
		DeleteEventGroupFn: func(_ context.Context, _, _ uuid.UUID) error { return nil },
	}, nil, "")
	h.DeleteEventGroupHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestUpdateEventGroupRegistrationStatusHandler_StoreErrors(t *testing.T) {
	gid := uuid.New()
	uid := uuid.New()
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"forbidden", store.ErrForbidden, http.StatusForbidden},
		{"teams", store.ErrOpenRegistrationTeams, http.StatusBadRequest},
		{"not found", pgx.ErrNoRows, http.StatusNotFound},
		{"other", errors.New("fail"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test_util.NewGinContext(http.MethodPatch, "/events/x/registration")
			test_util.WithUserIDString(c, uid)
			c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
			c.Request.Body = io.NopCloser(strings.NewReader(`{"registration_open":true}`))
			c.Request.Header.Set("Content-Type", "application/json")
			h := newTestHandler(t, &store.MockStore{
				SetEventGroupRegistrationOpenFn: func(_ context.Context, inG, inO uuid.UUID, open bool) error {
					assert.True(t, open)
					return tc.err
				},
			}, nil, "")
			h.UpdateEventGroupRegistrationStatusHandler(c)
			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUpdateEventGroupRegistrationStatusHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPatch, "/events/x/registration")
	c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"registration_open":true}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UpdateEventGroupRegistrationStatusHandler(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateEventGroupRegistrationStatusHandler_InvalidGroupID(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPatch, "/events/bad/registration")
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: "bad"}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"registration_open":true}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UpdateEventGroupRegistrationStatusHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateEventGroupRegistrationStatusHandler_BadJSON(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodPatch, "/events/x/registration")
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.UpdateEventGroupRegistrationStatusHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateEventGroupRegistrationStatusHandler_Success(t *testing.T) {
	c, _ := test_util.NewGinContext(http.MethodPatch, "/events/x/registration")
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"registration_open":false}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{
		SetEventGroupRegistrationOpenFn: func(_ context.Context, _, _ uuid.UUID, open bool) error {
			assert.False(t, open)
			return nil
		},
	}, nil, "")
	h.UpdateEventGroupRegistrationStatusHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestCreateTeamsHandler_StoreErrors(t *testing.T) {
	gid := uuid.New()
	uid := uuid.New()
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"forbidden", store.ErrForbidden, http.StatusForbidden},
		{"already", store.ErrTeamsAlreadyCreated, http.StatusBadRequest},
		{"not found", pgx.ErrNoRows, http.StatusNotFound},
		{"other", errors.New("fail"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test_util.NewGinContext(http.MethodPost, "/events/x/teams")
			test_util.WithUserIDString(c, uid)
			c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
			h := newTestHandler(t, &store.MockStore{
				CreateTeamsForGroupFn: func(_ context.Context, inG, inO uuid.UUID) error { return tc.err },
			}, nil, "")
			h.CreateTeamsHandler(c)
			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestCreateTeamsHandler_Success(t *testing.T) {
	c, _ := test_util.NewGinContext(http.MethodPost, "/events/x/teams")
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
	h := newTestHandler(t, &store.MockStore{
		CreateTeamsForGroupFn: func(_ context.Context, _, _ uuid.UUID) error { return nil },
	}, nil, "")
	h.CreateTeamsHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestDeleteTeamsAndOpenRegistrationHandler_StoreErrors(t *testing.T) {
	gid := uuid.New()
	uid := uuid.New()
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"forbidden", store.ErrForbidden, http.StatusForbidden},
		{"no teams", store.ErrTeamsNotCreated, http.StatusBadRequest},
		{"not found", pgx.ErrNoRows, http.StatusNotFound},
		{"other", errors.New("fail"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test_util.NewGinContext(http.MethodDelete, "/events/x/teams")
			test_util.WithUserIDString(c, uid)
			c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
			h := newTestHandler(t, &store.MockStore{
				DeleteTeamsAndOpenRegistrationFn: func(_ context.Context, _, _ uuid.UUID) error { return tc.err },
			}, nil, "")
			h.DeleteTeamsAndOpenRegistrationHandler(c)
			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestDeleteTeamsAndOpenRegistrationHandler_Success(t *testing.T) {
	c, _ := test_util.NewGinContext(http.MethodDelete, "/events/x/teams")
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
	h := newTestHandler(t, &store.MockStore{
		DeleteTeamsAndOpenRegistrationFn: func(_ context.Context, _, _ uuid.UUID) error { return nil },
	}, nil, "")
	h.DeleteTeamsAndOpenRegistrationHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestUpsertMyRegistrationHandler_Validation(t *testing.T) {
	t.Run("invalid event id", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/registrations/bad/me")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "eventId", Value: "nope"}}
		c.Request.Body = io.NopCloser(strings.NewReader(`{"can_substitute":true,"can_lobby_host":false,"duo_request":""}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertMyRegistrationHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("bad json", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/registrations/x/me")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "eventId", Value: uuid.NewString()}}
		c.Request.Body = io.NopCloser(strings.NewReader(`x`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertMyRegistrationHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpsertMyRegistrationHandler_StoreErrors(t *testing.T) {
	eid := uuid.New()
	uid := uuid.New()
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"closed", store.ErrRegistrationClosed, http.StatusBadRequest},
		{"incomplete profile", store.ErrUserGameProfileIncomplete, http.StatusBadRequest},
		{"not found", pgx.ErrNoRows, http.StatusNotFound},
		{"other", errors.New("fail"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test_util.NewGinContext(http.MethodPut, "/registrations/x/me")
			test_util.WithUserIDString(c, uid)
			c.Params = gin.Params{{Key: "eventId", Value: eid.String()}}
			c.Request.Body = io.NopCloser(strings.NewReader(`{"can_substitute":false,"can_lobby_host":true,"duo_request":"x"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			h := newTestHandler(t, &store.MockStore{
				UpsertRegistrationForEventFn: func(_ context.Context, inE, inU uuid.UUID, sub, host bool, duo *string) error {
					assert.Equal(t, eid, inE)
					assert.Equal(t, uid, inU)
					assert.False(t, sub)
					assert.True(t, host)
					require.NotNil(t, duo)
					assert.Equal(t, "x", *duo)
					return tc.err
				},
			}, nil, "")
			h.UpsertMyRegistrationHandler(c)
			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUpsertMyRegistrationHandler_SuccessTrimsDuo(t *testing.T) {
	eid := uuid.New()
	uid := uuid.New()
	c, _ := test_util.NewGinContext(http.MethodPut, "/registrations/x/me")
	test_util.WithUserIDString(c, uid)
	c.Params = gin.Params{{Key: "eventId", Value: eid.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"can_substitute":true,"can_lobby_host":false,"duo_request":"  "}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{
		UpsertRegistrationForEventFn: func(_ context.Context, _, _ uuid.UUID, _, _ bool, duo *string) error {
			assert.Nil(t, duo)
			return nil
		},
	}, nil, "")
	h.UpsertMyRegistrationHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestUpsertMyGroupRegistrationsHandler_Validation(t *testing.T) {
	t.Run("invalid group id", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/registrations/group/bad/me")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "groupId", Value: "bad"}}
		c.Request.Body = io.NopCloser(strings.NewReader(`{"events":[{"event_id":"` + uuid.NewString() + `","can_substitute":true,"can_lobby_host":false}]}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertMyGroupRegistrationsHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("bad json", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/registrations/group/x/me")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
		c.Request.Body = io.NopCloser(strings.NewReader(`not-json`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertMyGroupRegistrationsHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty events", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/registrations/group/x/me")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
		c.Request.Body = io.NopCloser(strings.NewReader(`{"events":[]}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertMyGroupRegistrationsHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid event id", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodPut, "/registrations/group/x/me")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
		c.Request.Body = io.NopCloser(strings.NewReader(`{"events":[{"event_id":"not-a-uuid","can_substitute":true,"can_lobby_host":false}]}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertMyGroupRegistrationsHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("duplicate event id", func(t *testing.T) {
		eventID := uuid.NewString()
		c, w := test_util.NewGinContext(http.MethodPut, "/registrations/group/x/me")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "groupId", Value: uuid.NewString()}}
		c.Request.Body = io.NopCloser(strings.NewReader(`{"events":[{"event_id":"` + eventID + `","can_substitute":true,"can_lobby_host":false},{"event_id":"` + eventID + `","can_substitute":false,"can_lobby_host":true}]}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.UpsertMyGroupRegistrationsHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpsertMyGroupRegistrationsHandler_StoreErrors(t *testing.T) {
	groupID := uuid.New()
	userID := uuid.New()
	eventID := uuid.New()

	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"invalid registration", store.ErrInvalidRegistration, http.StatusBadRequest},
		{"registration closed", store.ErrRegistrationClosed, http.StatusBadRequest},
		{"incomplete profile", store.ErrUserGameProfileIncomplete, http.StatusBadRequest},
		{"invalid event", store.ErrEventNotFound, http.StatusBadRequest},
		{"group not found", pgx.ErrNoRows, http.StatusNotFound},
		{"other", errors.New("fail"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test_util.NewGinContext(http.MethodPut, "/registrations/group/x/me")
			test_util.WithUserIDString(c, userID)
			c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
			c.Request.Body = io.NopCloser(strings.NewReader(`{"duo_request":" duo ","events":[{"event_id":"` + eventID.String() + `","can_substitute":true,"can_lobby_host":false}]}`))
			c.Request.Header.Set("Content-Type", "application/json")

			h := newTestHandler(t, &store.MockStore{
				UpsertRegistrationsForGroupFn: func(_ context.Context, inGroup, inUser uuid.UUID, regs []store.RegistrationUpsertItem, duo *string) error {
					assert.Equal(t, groupID, inGroup)
					assert.Equal(t, userID, inUser)
					require.Len(t, regs, 1)
					assert.Equal(t, eventID, regs[0].EventID)
					assert.True(t, regs[0].CanSubstitute)
					assert.False(t, regs[0].CanLobbyHost)
					require.NotNil(t, duo)
					assert.Equal(t, "duo", *duo)
					return tc.err
				},
			}, nil, "")
			h.UpsertMyGroupRegistrationsHandler(c)
			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestUpsertMyGroupRegistrationsHandler_Success(t *testing.T) {
	groupID := uuid.New()
	userID := uuid.New()
	eventID := uuid.New()

	c, _ := test_util.NewGinContext(http.MethodPut, "/registrations/group/x/me")
	test_util.WithUserIDString(c, userID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"duo_request":"  ","events":[{"event_id":"` + eventID.String() + `","can_substitute":false,"can_lobby_host":true}]}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := newTestHandler(t, &store.MockStore{
		UpsertRegistrationsForGroupFn: func(_ context.Context, inGroup, inUser uuid.UUID, regs []store.RegistrationUpsertItem, duo *string) error {
			assert.Equal(t, groupID, inGroup)
			assert.Equal(t, userID, inUser)
			require.Len(t, regs, 1)
			assert.Equal(t, eventID, regs[0].EventID)
			assert.False(t, regs[0].CanSubstitute)
			assert.True(t, regs[0].CanLobbyHost)
			assert.Nil(t, duo)
			return nil
		},
	}, nil, "")
	h.UpsertMyGroupRegistrationsHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestDeleteRegistrationHandler_Validation(t *testing.T) {
	t.Run("invalid event id", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodDelete, "/registrations/bad/x")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "eventId", Value: "bad"}, {Key: "userId", Value: uuid.NewString()}}
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.DeleteRegistrationHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("invalid user id param", func(t *testing.T) {
		c, w := test_util.NewGinContext(http.MethodDelete, "/registrations/x/baduser")
		test_util.WithUserIDString(c, uuid.New())
		c.Params = gin.Params{{Key: "eventId", Value: uuid.NewString()}, {Key: "userId", Value: "nope"}}
		h := newTestHandler(t, &store.MockStore{}, nil, "")
		h.DeleteRegistrationHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDeleteRegistrationHandler_StoreErrors(t *testing.T) {
	eid := uuid.New()
	actor := uuid.New()
	target := uuid.New()
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"forbidden", store.ErrForbidden, http.StatusForbidden},
		{"closed", store.ErrRegistrationClosed, http.StatusBadRequest},
		{"reg missing", store.ErrRegistrationNotFound, http.StatusNotFound},
		{"event missing", pgx.ErrNoRows, http.StatusNotFound},
		{"other", errors.New("fail"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test_util.NewGinContext(http.MethodDelete, "/registrations/x/y")
			test_util.WithUserIDString(c, actor)
			c.Params = gin.Params{{Key: "eventId", Value: eid.String()}, {Key: "userId", Value: target.String()}}
			h := newTestHandler(t, &store.MockStore{
				DeleteRegistrationForEventFn: func(_ context.Context, inE, inT, inA uuid.UUID) error {
					assert.Equal(t, eid, inE)
					assert.Equal(t, target, inT)
					assert.Equal(t, actor, inA)
					return tc.err
				},
			}, nil, "")
			h.DeleteRegistrationHandler(c)
			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestDeleteRegistrationHandler_SelfWithNoUserIDParam(t *testing.T) {
	eid := uuid.New()
	uid := uuid.New()
	c, _ := test_util.NewGinContext(http.MethodDelete, "/registrations/x/me")
	test_util.WithUserIDString(c, uid)
	c.Params = gin.Params{{Key: "eventId", Value: eid.String()}}
	h := newTestHandler(t, &store.MockStore{
		DeleteRegistrationForEventFn: func(_ context.Context, inE, inT, inA uuid.UUID) error {
			assert.Equal(t, uid, inT)
			return nil
		},
	}, nil, "")
	h.DeleteRegistrationHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}
