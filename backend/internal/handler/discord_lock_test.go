package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/discord"
	"github.com/KatieSuth/MatchmakerAPI/internal/handler"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type slogCapture struct {
	mu       sync.Mutex
	levels   []slog.Level
	messages []string
}

func (s *slogCapture) Enabled(context.Context, slog.Level) bool { return true }

func (s *slogCapture) Handle(_ context.Context, r slog.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.levels = append(s.levels, r.Level)
	s.messages = append(s.messages, r.Message)
	return nil
}

func (s *slogCapture) WithAttrs([]slog.Attr) slog.Handler { return s }
func (s *slogCapture) WithGroup(string) slog.Handler      { return s }

func (s *slogCapture) hasLevel(level slog.Level) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, l := range s.levels {
		if l == level {
			return true
		}
	}
	return false
}

func captureSlog(t *testing.T) *slogCapture {
	t.Helper()
	cap := &slogCapture{}
	prev := slog.Default()
	slog.SetDefault(slog.New(cap))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return cap
}

func handlerTxStore(t *testing.T) *store.PostgresStore {
	t.Helper()
	pool := test_util.GetTestPool(t)
	t.Cleanup(func() { pool.Close() })
	tx, err := pool.Begin(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	return store.NewPostgresStoreFromTx(tx)
}

func createHandlerTestUser(t *testing.T, s *store.PostgresStore) model.User {
	t.Helper()
	suffix := uuid.NewString()
	user, err := s.CreateNewUser(context.Background(), model.DiscordUser{
		ID:       "discord-h-" + suffix,
		Username: "htest-" + suffix,
		Avatar:   "avatar",
	}, nil)
	require.NoError(t, err)
	return user
}

func firstHandlerMode(t *testing.T, s *store.PostgresStore) model.GameMode {
	t.Helper()
	games, err := s.GetSystemGames(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, games)
	modes, err := s.GetGameModes(context.Background(), games[0].ID)
	require.NoError(t, err)
	for _, m := range modes {
		if m.Duration > 0 {
			return m
		}
	}
	t.Fatal("need a game mode with positive duration")
	return model.GameMode{}
}

func lockedGroup(t *testing.T, s *store.PostgresStore, host model.User, guilds []model.DiscordGuild) (uuid.UUID, uuid.UUID) {
	t.Helper()
	mode := firstHandlerMode(t, s)
	start := time.Now().UTC().Add(24 * time.Hour)
	groupID, err := s.CreateEventGroupWithEvents(context.Background(), host.ID, mode.ID, 0, true, "AMER", "balanced", "Locked Scrims", start, 1, guilds)
	require.NoError(t, err)
	detail, err := s.GetEventGroupDetail(context.Background(), groupID, host.ID)
	require.NoError(t, err)
	require.NotEmpty(t, detail.Events)
	return groupID, detail.Events[0].ID
}

func accessJSON(t *testing.T, w *bytes.Buffer) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal(w.Bytes(), &out))
	return out
}

func TestGetEventGroupAccessHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/events/x/access")
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetEventGroupAccessHandler_InvalidGroupID(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/events/bad/access")
	test_util.WithUserIDString(c, uuid.New())
	c.Params = gin.Params{{Key: "groupId", Value: "bad"}}
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetEventGroupAccessHandler_OwnerAllowedWithoutDiscord(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	groupID, _ := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})

	logs := captureSlog(t)
	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(c, host.ID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{listErr: errors.New("should not call discord")})
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, logs.hasLevel(slog.LevelError))
}

func TestGetEventGroupAccessHandler_UnrestrictedAllowsNonMember(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	groupID, _ := lockedGroup(t, s, host, nil)

	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{listErr: errors.New("should not call discord")})
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetEventGroupAccessHandler_MemberAllowed(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	groupID, _ := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}, {ID: "g2", Name: "Beta"}})

	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{
		guildsByUser: map[uuid.UUID][]model.DiscordGuild{
			guest.ID: {{ID: "g2", Name: "Beta"}},
		},
	})
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetEventGroupAccessHandler_NonMemberForbiddenWarn(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	groupID, _ := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})
	logs := captureSlog(t)

	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{
		guildsByUser: map[uuid.UUID][]model.DiscordGuild{
			guest.ID: {{ID: "other", Name: "Other"}},
		},
	})
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
	body := accessJSON(t, w.Body)
	assert.Equal(t, "error", body["status"])
	details := body["details"].(map[string]any)
	assert.Equal(t, "discord_guild_restricted", details["code"])
	assert.Equal(t, "Locked Scrims", details["event_title"])
	assert.Equal(t, true, details["event_named"])
	assert.True(t, logs.hasLevel(slog.LevelWarn))
	assert.False(t, logs.hasLevel(slog.LevelError))
}

func TestGetEventGroupAccessHandler_UnnamedEventTitleIsGame(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	mode := firstHandlerMode(t, s)
	games, err := s.GetSystemGames(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, games)
	start := time.Now().UTC().Add(24 * time.Hour)
	groupID, err := s.CreateEventGroupWithEvents(context.Background(), host.ID, mode.ID, 0, true, "AMER", "balanced", "", start, 1, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})
	require.NoError(t, err)

	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{})
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
	details := accessJSON(t, w.Body)["details"].(map[string]any)
	assert.Equal(t, false, details["event_named"])
	assert.Equal(t, games[0].Name, details["event_title"])
}

func TestGetEventGroupAccessHandler_DiscordErrorForbiddenErrorLog(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	groupID, _ := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})
	logs := captureSlog(t)

	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{listErr: discord.ErrMissingGrant})
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.True(t, logs.hasLevel(slog.LevelError))
}

func TestGetEventGroupAccessHandler_CanceledDoesNotLock(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	groupID, _ := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})

	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{listErr: context.Canceled})
	h.GetEventGroupAccessHandler(c)
	assert.NotContains(t, w.Body.String(), "discord_guild_restricted")
}

func TestGetEventGroupHandler_DiscordRestricted(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	groupID, _ := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})

	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String())
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{})
	h.GetEventGroupHandler(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpsertMyRegistrationHandler_DiscordRestricted(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	_, eventID := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})

	c, w := test_util.NewGinContext(http.MethodPut, "/registrations/"+eventID.String()+"/me")
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "eventId", Value: eventID.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"can_substitute":true,"can_lobby_host":false,"duo_request":""}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{})
	h.UpsertMyRegistrationHandler(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpsertMyGroupRegistrationsHandler_DiscordRestricted(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	groupID, eventID := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})

	c, w := test_util.NewGinContext(http.MethodPut, "/registrations/group/"+groupID.String()+"/me")
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"duo_request":"","events":[{"event_id":"` + eventID.String() + `","can_substitute":true,"can_lobby_host":false}]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{})
	h.UpsertMyGroupRegistrationsHandler(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteRegistrationHandler_SkipsDiscord(t *testing.T) {
	eventID := uuid.New()
	userID := uuid.New()
	called := false
	ms := &store.MockStore{
		DeleteRegistrationForEventFn: func(_ context.Context, _, _, _ uuid.UUID) error {
			called = true
			return nil
		},
		ListEventGroupDiscordGuildsFn: func(_ context.Context, _ uuid.UUID) ([]model.DiscordGuild, error) {
			t.Fatal("unregister must not check discord guilds")
			return nil, nil
		},
	}
	c, _ := test_util.NewGinContext(http.MethodDelete, "/registrations/"+eventID.String()+"/me")
	test_util.WithUserIDString(c, userID)
	c.Params = gin.Params{{Key: "eventId", Value: eventID.String()}}
	h := newTestHandler(t, ms, nil, "")
	h.DeleteRegistrationHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.True(t, called)
}

func TestListMyDiscordGuildsHandler_Success(t *testing.T) {
	userID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/discord/guilds")
	test_util.WithUserIDString(c, userID)
	h := newTestHandlerWithDiscord(t, &store.MockStore{}, nil, "", "", &fakeDiscordAPI{
		guilds: []model.DiscordGuild{{ID: "g1", Name: "Alpha"}},
	})
	h.ListMyDiscordGuildsHandler(c)
	assert.Equal(t, http.StatusOK, w.Code)
	body := accessJSON(t, w.Body)
	guilds := body["guilds"].([]any)
	require.Len(t, guilds, 1)
	assert.Equal(t, "g1", guilds[0].(map[string]any)["id"])
}

func TestListMyDiscordGuildsHandler_GrantExpired(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/discord/guilds")
	test_util.WithUserIDString(c, uuid.New())
	h := newTestHandlerWithDiscord(t, &store.MockStore{}, nil, "", "", &fakeDiscordAPI{listErr: discord.ErrInvalidGrant})
	h.ListMyDiscordGuildsHandler(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateEventHandler_DiscordGuildNotMember(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	storeCalled := false
	body := `{
	  "game_mode_id": "` + gameModeID.String() + `",
	  "region": "AMER",
	  "start_time": "` + startTimeHoursFromNow(48) + `",
	  "sub_min": 0,
	  "games_to_run": 1,
	  "registration_open": true,
	  "discord_guild_ids": ["g1"]
	}`
	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandlerWithDiscord(t, &store.MockStore{
		CreateEventGroupWithEventsFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int32, _ bool, _ string, _ string, _ string, _ time.Time, _ int32, _ []model.DiscordGuild) (uuid.UUID, error) {
			storeCalled = true
			return uuid.Nil, nil
		},
	}, nil, "", "", &fakeDiscordAPI{guilds: []model.DiscordGuild{{ID: "other", Name: "Other"}}})
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, storeCalled)
}

func TestCreateEventHandler_DiscordGuildsSnapshot(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	var got []model.DiscordGuild
	body := `{
	  "game_mode_id": "` + gameModeID.String() + `",
	  "region": "AMER",
	  "start_time": "` + startTimeHoursFromNow(48) + `",
	  "sub_min": 0,
	  "games_to_run": 1,
	  "registration_open": true,
	  "discord_guild_ids": ["g1"]
	}`
	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandlerWithDiscord(t, &store.MockStore{
		CreateEventGroupWithEventsFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int32, _ bool, _ string, _ string, _ string, _ time.Time, _ int32, guilds []model.DiscordGuild) (uuid.UUID, error) {
			got = guilds
			return uuid.New(), nil
		},
	}, nil, "", "", &fakeDiscordAPI{guilds: []model.DiscordGuild{{ID: "g1", Name: "Alpha"}}})
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusCreated, w.Code)
	require.Len(t, got, 1)
	assert.Equal(t, "Alpha", got[0].Name)
}

func TestUpdateEventGroupSettingsHandler_ClearsDiscordGuilds(t *testing.T) {
	gid := uuid.New()
	uid := uuid.New()
	var got []model.DiscordGuild
	c, _ := test_util.NewGinContext(http.MethodPatch, "/events/x")
	test_util.WithUserIDString(c, uid)
	c.Params = gin.Params{{Key: "groupId", Value: gid.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"region":"EU","sub_min":2,"sort_logic":"ranked","registration_open":false,"discord_guild_ids":[],"events":[{"event_id":"33333333-3333-3333-3333-333333333333","start_time":"2099-03-01T15:30:00Z","game_mode_id":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"}]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{
		UpdateEventGroupSettingsFn: func(_ context.Context, _, _ uuid.UUID, _ string, _ int32, _ string, _ bool, _ string, _ []store.GroupEventUpdate, guilds []model.DiscordGuild) error {
			got = guilds
			return nil
		},
	}, nil, "")
	h.UpdateEventGroupSettingsHandler(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.Empty(t, got)
}

func TestClearDiscordLockThenGuestAccessAllowed(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	groupID, _ := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})
	detail, err := s.GetEventGroupDetail(context.Background(), groupID, host.ID)
	require.NoError(t, err)
	require.NotEmpty(t, detail.Events)
	ev := detail.Events[0]

	refusing := &fakeDiscordAPI{listErr: errors.New("should not call discord after lock is cleared")}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", refusing)

	patch, err := json.Marshal(map[string]any{
		"region":            detail.Region,
		"sub_min":           detail.SubMin,
		"sort_logic":        detail.SortLogic,
		"registration_open": true,
		"name":              detail.Name,
		"discord_guild_ids": []string{},
		"events": []map[string]any{{
			"event_id":     ev.ID.String(),
			"start_time":   ev.StartTime.UTC().Format(time.RFC3339),
			"game_mode_id": ev.GameModeID.String(),
		}},
	})
	require.NoError(t, err)
	cPatch, _ := test_util.NewGinContext(http.MethodPatch, "/events/"+groupID.String())
	test_util.WithUserIDString(cPatch, host.ID)
	cPatch.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	cPatch.Request.Body = io.NopCloser(bytes.NewReader(patch))
	cPatch.Request.Header.Set("Content-Type", "application/json")
	h.UpdateEventGroupSettingsHandler(cPatch)
	assert.Equal(t, http.StatusNoContent, cPatch.Writer.Status())

	remaining, err := s.ListEventGroupDiscordGuilds(context.Background(), groupID)
	require.NoError(t, err)
	assert.Empty(t, remaining)

	cAccess, wAccess := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(cAccess, guest.ID)
	cAccess.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h.GetEventGroupAccessHandler(cAccess)
	assert.Equal(t, http.StatusOK, wAccess.Code)

	cGet, wGet := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String())
	test_util.WithUserIDString(cGet, guest.ID)
	cGet.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h.GetEventGroupHandler(cGet)
	assert.Equal(t, http.StatusOK, wGet.Code)
}

func TestUpsertMyRegistrationHandler_StaleMembershipAfterAccessAllowed(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	groupID, eventID := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})
	fake := &fakeDiscordAPI{
		guildsByUser: map[uuid.UUID][]model.DiscordGuild{
			guest.ID: {{ID: "g1", Name: "Alpha"}},
		},
	}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", fake)

	cAccess, wAccess := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(cAccess, guest.ID)
	cAccess.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h.GetEventGroupAccessHandler(cAccess)
	assert.Equal(t, http.StatusOK, wAccess.Code)

	cGet, wGet := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String())
	test_util.WithUserIDString(cGet, guest.ID)
	cGet.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h.GetEventGroupHandler(cGet)
	assert.Equal(t, http.StatusOK, wGet.Code)

	fake.guildsByUser[guest.ID] = []model.DiscordGuild{{ID: "other", Name: "Other"}}

	cReg, wReg := test_util.NewGinContext(http.MethodPut, "/registrations/"+eventID.String()+"/me")
	test_util.WithUserIDString(cReg, guest.ID)
	cReg.Params = gin.Params{{Key: "eventId", Value: eventID.String()}}
	cReg.Request.Body = io.NopCloser(strings.NewReader(`{"can_substitute":true,"can_lobby_host":false,"duo_request":""}`))
	cReg.Request.Header.Set("Content-Type", "application/json")
	h.UpsertMyRegistrationHandler(cReg)
	assert.Equal(t, http.StatusForbidden, wReg.Code)
	body := accessJSON(t, wReg.Body)
	details := body["details"].(map[string]any)
	assert.Equal(t, "discord_guild_restricted", details["code"])
}

func TestGetEventGroupAccessHandler_NotFound(t *testing.T) {
	s := handlerTxStore(t)
	guest := createHandlerTestUser(t, s)
	missing := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+missing.String()+"/access")
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "groupId", Value: missing.String()}}
	h := newTestHandlerWithDiscord(t, s, nil, "", "", &fakeDiscordAPI{})
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetEventGroupAccessHandler_ListGuildsStoreError(t *testing.T) {
	userID := uuid.New()
	groupID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(c, userID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandler(t, &store.MockStore{
		GetEventGroupAccessMetaFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, string, bool, error) {
			return uuid.New(), "t", false, nil
		},
		ListEventGroupDiscordGuildsFn: func(_ context.Context, _ uuid.UUID) ([]model.DiscordGuild, error) {
			return nil, errors.New("db down")
		},
	}, nil, "")
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRequireEventDiscordAllowed_StoreError(t *testing.T) {
	userID := uuid.New()
	groupID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(c, userID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandler(t, &store.MockStore{
		GetEventGroupAccessMetaFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, string, bool, error) {
			return uuid.Nil, "", false, errors.New("db down")
		},
	}, nil, "")
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpsertMyRegistrationHandler_EventNotFoundForDiscord(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodPut, "/registrations/"+eventID.String()+"/me")
	test_util.WithUserIDString(c, userID)
	c.Params = gin.Params{{Key: "eventId", Value: eventID.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"can_substitute":true,"can_lobby_host":false}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{
		EventGroupIDByEventIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.Nil, store.ErrEventNotFound
		},
	}, nil, "")
	h.UpsertMyRegistrationHandler(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpsertMyRegistrationHandler_EventResolveError(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodPut, "/registrations/"+eventID.String()+"/me")
	test_util.WithUserIDString(c, userID)
	c.Params = gin.Params{{Key: "eventId", Value: eventID.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"can_substitute":true,"can_lobby_host":false}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{
		EventGroupIDByEventIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.Nil, errors.New("db down")
		},
	}, nil, "")
	h.UpsertMyRegistrationHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateLobbyJoinCodeHandler_LobbyNotFoundForDiscord(t *testing.T) {
	userID := uuid.New()
	lobbyID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodPatch, "/lobbies/"+lobbyID.String()+"/join-code")
	test_util.WithUserIDString(c, userID)
	c.Params = gin.Params{{Key: "lobbyId", Value: lobbyID.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"join_code":"abc"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{
		EventGroupIDByLobbyIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.Nil, store.ErrLobbyNotFound
		},
	}, nil, "")
	h.UpdateLobbyJoinCodeHandler(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateLobbyJoinCodeHandler_LobbyResolveError(t *testing.T) {
	userID := uuid.New()
	lobbyID := uuid.New()
	c, w := test_util.NewGinContext(http.MethodPatch, "/lobbies/"+lobbyID.String()+"/join-code")
	test_util.WithUserIDString(c, userID)
	c.Params = gin.Params{{Key: "lobbyId", Value: lobbyID.String()}}
	c.Request.Body = io.NopCloser(strings.NewReader(`{"join_code":"abc"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{
		EventGroupIDByLobbyIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.Nil, errors.New("db down")
		},
	}, nil, "")
	h.UpdateLobbyJoinCodeHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateEventHandler_DiscordNotConfigured(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	body := `{
	  "game_mode_id": "` + gameModeID.String() + `",
	  "region": "AMER",
	  "start_time": "` + startTimeHoursFromNow(48) + `",
	  "sub_min": 0,
	  "games_to_run": 1,
	  "registration_open": true,
	  "discord_guild_ids": ["g1"]
	}`
	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	handler.SetDiscordAPIForTest(h, nil)
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateEventHandler_DiscordGuildsLoadError(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	body := `{
	  "game_mode_id": "` + gameModeID.String() + `",
	  "region": "AMER",
	  "start_time": "` + startTimeHoursFromNow(48) + `",
	  "sub_min": 0,
	  "games_to_run": 1,
	  "registration_open": true,
	  "discord_guild_ids": ["g1"]
	}`
	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandlerWithDiscord(t, &store.MockStore{}, nil, "", "", &fakeDiscordAPI{listErr: errors.New("timeout")})
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateEventHandler_DedupesGuildIDs(t *testing.T) {
	userID := uuid.New()
	gameModeID := uuid.New()
	var got []model.DiscordGuild
	body := `{
	  "game_mode_id": "` + gameModeID.String() + `",
	  "region": "AMER",
	  "start_time": "` + startTimeHoursFromNow(48) + `",
	  "sub_min": 0,
	  "games_to_run": 1,
	  "registration_open": true,
	  "discord_guild_ids": ["g1", "g1", ""]
	}`
	c, w := test_util.NewGinContext(http.MethodPost, "/events")
	test_util.WithUserIDString(c, userID)
	c.Request.Body = io.NopCloser(strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h := newTestHandlerWithDiscord(t, &store.MockStore{
		CreateEventGroupWithEventsFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int32, _ bool, _ string, _ string, _ string, _ time.Time, _ int32, guilds []model.DiscordGuild) (uuid.UUID, error) {
			got = guilds
			return uuid.New(), nil
		},
	}, nil, "", "", &fakeDiscordAPI{guilds: []model.DiscordGuild{{ID: "g1", Name: "Alpha"}}})
	h.CreateEventHandler(c)
	assert.Equal(t, http.StatusCreated, w.Code)
	require.Len(t, got, 1)
}

func TestListMyDiscordGuildsHandler_Unauthorized(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/discord/guilds")
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	h.ListMyDiscordGuildsHandler(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListMyDiscordGuildsHandler_NilDiscord(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/discord/guilds")
	test_util.WithUserIDString(c, uuid.New())
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	handler.SetDiscordAPIForTest(h, nil)
	h.ListMyDiscordGuildsHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListMyDiscordGuildsHandler_TransientError(t *testing.T) {
	c, w := test_util.NewGinContext(http.MethodGet, "/users/me/discord/guilds")
	test_util.WithUserIDString(c, uuid.New())
	h := newTestHandlerWithDiscord(t, &store.MockStore{}, nil, "", "", &fakeDiscordAPI{listErr: errors.New("429")})
	h.ListMyDiscordGuildsHandler(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDiscordRestrictionErrorHelpers(t *testing.T) {
	cause := errors.New("boom")
	err := handler.DiscordRestrictionErrorForTest(cause)
	assert.Equal(t, "boom", err.Error())
	assert.ErrorIs(t, err, cause)

	none := handler.DiscordRestrictionErrorForTest(nil)
	assert.ErrorIs(t, none, handler.ErrDiscordGuildRestricted)
	assert.Equal(t, handler.ErrDiscordGuildRestricted.Error(), none.Error())

	c, w := test_util.NewGinContext(http.MethodGet, "/x")
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	handler.WriteDiscordGuildRestrictionForTest(h, c, uuid.New(), uuid.New())
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetEventGroupAccessHandler_DiscordNotConfigured(t *testing.T) {
	s := handlerTxStore(t)
	host := createHandlerTestUser(t, s)
	guest := createHandlerTestUser(t, s)
	groupID, _ := lockedGroup(t, s, host, []model.DiscordGuild{{ID: "g1", Name: "Alpha"}})
	c, w := test_util.NewGinContext(http.MethodGet, "/events/"+groupID.String()+"/access")
	test_util.WithUserIDString(c, guest.ID)
	c.Params = gin.Params{{Key: "groupId", Value: groupID.String()}}
	h := newTestHandler(t, s, nil, "")
	handler.SetDiscordAPIForTest(h, nil)
	h.GetEventGroupAccessHandler(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
