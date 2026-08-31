package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventGroupDiscordGuilds_ReplaceListAndEmpty(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	start := time.Now().UTC().Add(24 * time.Hour)

	groupID, err := s.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "", start, 1, []model.DiscordGuild{
		{ID: "111", Name: "Alpha"},
		{ID: "222", Name: "Beta"},
	})
	require.NoError(t, err)

	got, err := s.ListEventGroupDiscordGuilds(ctx, groupID)
	require.NoError(t, err)
	require.Len(t, got, 2)
	byID := map[string]string{}
	for _, g := range got {
		byID[g.ID] = g.Name
	}
	assert.Equal(t, "Alpha", byID["111"])
	assert.Equal(t, "Beta", byID["222"])

	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	require.Len(t, detail.DiscordGuilds, 2)

	eventID := detail.Events[0].ID
	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, "", patchEventUpdates(eventID, mode.ID, start), []model.DiscordGuild{
		{ID: "333", Name: "Gamma"},
	})
	require.NoError(t, err)
	got, err = s.ListEventGroupDiscordGuilds(ctx, groupID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "333", got[0].ID)
	assert.Equal(t, "Gamma", got[0].Name)

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, "", patchEventUpdates(eventID, mode.ID, start), []model.DiscordGuild{})
	require.NoError(t, err)
	got, err = s.ListEventGroupDiscordGuilds(ctx, groupID)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestEventGroupDiscordGuilds_CascadeOnGroupDelete(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	start := time.Now().UTC().Add(24 * time.Hour)

	groupID, err := s.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "", start, 1, []model.DiscordGuild{
		{ID: "111", Name: "Alpha"},
	})
	require.NoError(t, err)

	err = s.DeleteEventGroup(ctx, groupID, host.ID)
	require.NoError(t, err)

	got, err := s.ListEventGroupDiscordGuilds(ctx, groupID)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestEventGroupDiscordGuilds_UniquePair(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	start := time.Now().UTC().Add(24 * time.Hour)

	_, err = s.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "", start, 1, []model.DiscordGuild{
		{ID: "111", Name: "Alpha"},
		{ID: "111", Name: "Alpha again"},
	})
	require.Error(t, err)
}

func TestGetEventGroupAccessMeta_UsesNameOrGameName(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	start := time.Now().UTC().Add(24 * time.Hour)

	namedID, err := s.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "Scrims", start, 1, nil)
	require.NoError(t, err)
	owner, title, named, err := s.GetEventGroupAccessMeta(ctx, namedID)
	require.NoError(t, err)
	assert.Equal(t, host.ID, owner)
	assert.Equal(t, "Scrims", title)
	assert.True(t, named)

	unnamedID, err := s.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "", start, 1, nil)
	require.NoError(t, err)
	_, title, named, err = s.GetEventGroupAccessMeta(ctx, unnamedID)
	require.NoError(t, err)
	assert.Equal(t, games[0].Name, title)
	assert.False(t, named)
}

func TestGetEventGroupAccessMeta_NotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	_, _, _, err := s.GetEventGroupAccessMeta(context.Background(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventGroupNotFound)
}

func TestEventGroupIDByEventID_Success(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	start := time.Now().UTC().Add(24 * time.Hour)
	groupID, err := s.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "", start, 1, nil)
	require.NoError(t, err)
	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	require.NotEmpty(t, detail.Events)
	got, err := s.EventGroupIDByEventID(ctx, detail.Events[0].ID)
	require.NoError(t, err)
	assert.Equal(t, groupID, got)
}

func TestEventGroupIDByEventID_NotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	_, err := s.EventGroupIDByEventID(context.Background(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrEventNotFound)
}

func TestEventGroupIDByLobbyID_Success(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	start := time.Now().UTC().Add(24 * time.Hour)
	groupID, err := s.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "", start, 1, nil)
	require.NoError(t, err)
	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	require.NotEmpty(t, detail.Events)
	lobbyID := insertLobbyForEvent(t, ctx, tx, detail.Events[0].ID, &host.ID)
	got, err := s.EventGroupIDByLobbyID(ctx, lobbyID)
	require.NoError(t, err)
	assert.Equal(t, groupID, got)
}

func TestEventGroupIDByLobbyID_NotFound(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	_, err := s.EventGroupIDByLobbyID(context.Background(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrLobbyNotFound)
}

func TestEventGroupDiscordGuilds_QueryErrors(t *testing.T) {
	s, _ := createEventTestStoreTx(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := s.ListEventGroupDiscordGuilds(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list event group discord guilds")

	_, _, _, err = s.GetEventGroupAccessMeta(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get event group access meta")

	_, err = s.EventGroupIDByEventID(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event group id by event")

	_, err = s.EventGroupIDByLobbyID(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event group id by lobby")
}

func TestReplaceEventGroupDiscordGuilds_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	mode := firstModeForGame(t, ctx, s, games[0].ID)
	start := time.Now().UTC().Add(24 * time.Hour)
	guilds := []model.DiscordGuild{{ID: "111", Name: "Alpha"}}

	t.Run("delete fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: tx, failOnExecCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)
		_, err := fs.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "", start, 1, guilds)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete event group discord guilds")
	})

	t.Run("insert fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: tx, failOnQueryRowCall: 3, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)
		_, err := fs.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "", start, 1, guilds)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insert event group discord guild")
	})
}
