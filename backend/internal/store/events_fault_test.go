package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type eventGroupFixture struct {
	s       *store.PostgresStore
	tx      db.DBTX
	ctx     context.Context
	hostID  uuid.UUID
	groupID uuid.UUID
	eventID uuid.UUID
	modeID  uuid.UUID
	gameID  uuid.UUID
}

func setupEventGroupFixture(t *testing.T) eventGroupFixture {
	t.Helper()
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	evStart := time.Now().UTC().Add(24 * time.Hour)
	groupID, eventID := insertEventFixture(t, ctx, tx, host.ID, modeID, evStart)

	return eventGroupFixture{
		s:       s,
		tx:      tx,
		ctx:     ctx,
		hostID:  host.ID,
		groupID: groupID,
		eventID: eventID,
		modeID:  modeID,
		gameID:  games[0].ID,
	}
}

func TestCreateEventGroupWithEvents_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")

	t.Run("get game mode fails", func(t *testing.T) {
		s, tx := createEventTestStoreTx(t)
		ctx := context.Background()
		host := createTestUser(t, ctx, s)
		faulty := &faultInjectTx{DBTX: tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := fs.CreateEventGroupWithEvents(ctx, host.ID, uuid.New(), 0, true, "AMER", "balanced", "", time.Now().UTC().Add(24*time.Hour), 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get game mode by id")
	})

	t.Run("create event group fails", func(t *testing.T) {
		s, tx := createEventTestStoreTx(t)
		ctx := context.Background()
		host := createTestUser(t, ctx, s)
		games, err := s.GetSystemGames(ctx)
		require.NoError(t, err)
		mode := firstModeForGame(t, ctx, s, games[0].ID)
		faulty := &faultInjectTx{DBTX: tx, failOnQueryRowCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err = fs.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "", time.Now().UTC().Add(24*time.Hour), 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create event group")
	})

	t.Run("create event fails", func(t *testing.T) {
		s, tx := createEventTestStoreTx(t)
		ctx := context.Background()
		host := createTestUser(t, ctx, s)
		games, err := s.GetSystemGames(ctx)
		require.NoError(t, err)
		mode := firstModeForGame(t, ctx, s, games[0].ID)
		faulty := &faultInjectTx{DBTX: tx, failOnExecCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err = fs.CreateEventGroupWithEvents(ctx, host.ID, mode.ID, 0, true, "AMER", "balanced", "", time.Now().UTC().Add(24*time.Hour), 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create event #1")
	})
}

func TestGetEventGroupDetail_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)

	t.Run("group detail lookup fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := fs.GetEventGroupDetail(fixture.ctx, fixture.groupID, fixture.hostID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get event group detail")
	})

	t.Run("events summary fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := fs.GetEventGroupDetail(fixture.ctx, fixture.groupID, fixture.hostID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get group events summary")
	})

	t.Run("registration lookup fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := fs.GetEventGroupDetail(fixture.ctx, fixture.groupID, fixture.hostID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get registration data for event")
	})

	t.Run("lobby load fails", func(t *testing.T) {
		insertLobbyForEvent(t, fixture.ctx, fixture.tx, fixture.eventID, &fixture.hostID)
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryCall: 3, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := fs.GetEventGroupDetail(fixture.ctx, fixture.groupID, fixture.hostID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "load lobbies for event")
	})
}

func TestUpdateEventGroupSettings_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)
	evStart := time.Now().UTC().Add(48 * time.Hour)
	updates := patchEventUpdates(fixture.eventID, fixture.modeID, evStart)

	t.Run("get event group fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpdateEventGroupSettings(fixture.ctx, fixture.groupID, fixture.hostID, "AMER", 0, "balanced", true, "", updates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get event group")
	})

	t.Run("count lobbies fails when opening registration", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpdateEventGroupSettings(fixture.ctx, fixture.groupID, fixture.hostID, "AMER", 0, "balanced", true, "", updates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "count lobbies by group")
	})

	t.Run("get events by group fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpdateEventGroupSettings(fixture.ctx, fixture.groupID, fixture.hostID, "AMER", 0, "balanced", false, "", updates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get events by group")
	})

	t.Run("get event group detail fails when opening registration", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 3, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpdateEventGroupSettings(fixture.ctx, fixture.groupID, fixture.hostID, "AMER", 0, "balanced", true, "", updates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get event group detail")
	})

	t.Run("get game mode fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 3, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpdateEventGroupSettings(fixture.ctx, fixture.groupID, fixture.hostID, "AMER", 0, "balanced", false, "", updates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get game mode")
	})

	t.Run("update event group settings fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 4, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpdateEventGroupSettings(fixture.ctx, fixture.groupID, fixture.hostID, "AMER", 0, "balanced", false, "", updates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "update event group settings")
	})

	t.Run("update event schedule fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnExecCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpdateEventGroupSettings(fixture.ctx, fixture.groupID, fixture.hostID, "AMER", 0, "balanced", false, "", updates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "update event schedule")
	})
}

func TestUpdateEventGroupSettings_UnknownEventID(t *testing.T) {
	fixture := setupEventGroupFixture(t)
	evStart := time.Now().UTC().Add(48 * time.Hour)

	err := fixture.s.UpdateEventGroupSettings(fixture.ctx, fixture.groupID, fixture.hostID, "AMER", 0, "balanced", true, "",
		patchEventUpdates(uuid.New(), fixture.modeID, evStart))
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidGroupEvents)
}

func TestUpdateEventGroupSettings_DuplicateEventInPayload(t *testing.T) {
	fixture := setupEventGroupFixture(t)
	evStart := time.Now().UTC().Add(48 * time.Hour)
	insertEventInGroupFixture(t, fixture.ctx, fixture.tx, fixture.groupID, fixture.modeID, evStart.Add(24*time.Hour))
	dup := patchEventUpdates(fixture.eventID, fixture.modeID, evStart)
	updates := append(dup, dup...)

	err := fixture.s.UpdateEventGroupSettings(fixture.ctx, fixture.groupID, fixture.hostID, "AMER", 0, "balanced", true, "", updates)
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidGroupEvents)
}

func TestUpdateEventGroupSettings_MismatchedEventCount(t *testing.T) {
	fixture := setupEventGroupFixture(t)
	evStart := time.Now().UTC().Add(48 * time.Hour)
	insertEventInGroupFixture(t, fixture.ctx, fixture.tx, fixture.groupID, fixture.modeID, evStart.Add(24*time.Hour))

	err := fixture.s.UpdateEventGroupSettings(fixture.ctx, fixture.groupID, fixture.hostID, "AMER", 0, "balanced", true, "",
		patchEventUpdates(fixture.eventID, fixture.modeID, evStart))
	require.Error(t, err)
	assert.ErrorIs(t, err, store.ErrInvalidGroupEvents)
}

func TestDeleteEventGroup_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)

	t.Run("get event group fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.DeleteEventGroup(fixture.ctx, fixture.groupID, fixture.hostID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get event group")
	})

	t.Run("delete event group fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnExecCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.DeleteEventGroup(fixture.ctx, fixture.groupID, fixture.hostID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete event group")
	})
}

func TestSetEventGroupRegistrationOpen_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)

	t.Run("get event group fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.SetEventGroupRegistrationOpen(fixture.ctx, fixture.groupID, fixture.hostID, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get event group")
	})

	t.Run("count lobbies fails when opening", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.SetEventGroupRegistrationOpen(fixture.ctx, fixture.groupID, fixture.hostID, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "count lobbies by group")
	})

	t.Run("set registration open fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.SetEventGroupRegistrationOpen(fixture.ctx, fixture.groupID, fixture.hostID, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "set registration_open")
	})
}

func TestCreateTeamsForGroup_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)
	for i := 0; i < 4; i++ {
		u := createTestUser(t, fixture.ctx, fixture.s)
		registerPlayerForEventWithProfile(t, fixture.ctx, fixture.tx, fixture.s, fixture.eventID, u.ID, fixture.gameID, false, false)
	}

	t.Run("get event group fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := fs.CreateTeamsForGroup(fixture.ctx, fixture.groupID, fixture.hostID, defaultMatchmakingSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get event group")
	})

	t.Run("count lobbies fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := fs.CreateTeamsForGroup(fixture.ctx, fixture.groupID, fixture.hostID, defaultMatchmakingSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "count lobbies by group")
	})

	t.Run("close registration fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 5, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := fs.CreateTeamsForGroup(fixture.ctx, fixture.groupID, fixture.hostID, defaultMatchmakingSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "close registration for group")
	})

	t.Run("persist lobby fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 6, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := fs.CreateTeamsForGroup(fixture.ctx, fixture.groupID, fixture.hostID, defaultMatchmakingSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create lobby")
	})
}

func TestDeleteTeamsForGroup_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)
	insertLobbyForEvent(t, fixture.ctx, fixture.tx, fixture.eventID, &fixture.hostID)

	t.Run("count lobbies fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.DeleteTeamsForGroup(fixture.ctx, fixture.groupID, fixture.hostID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "count lobbies by group")
	})

	t.Run("delete players fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnExecCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.DeleteTeamsForGroup(fixture.ctx, fixture.groupID, fixture.hostID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete players by group")
	})

	t.Run("delete lobbies fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnExecCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.DeleteTeamsForGroup(fixture.ctx, fixture.groupID, fixture.hostID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete lobbies by group")
	})
}

func TestUpsertRegistrationForEvent_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)
	joiner := createTestUser(t, fixture.ctx, fixture.s)
	createCompleteUserGameForRegistration(t, fixture.ctx, fixture.tx, fixture.s, joiner.ID, fixture.gameID)

	t.Run("get event fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationForEvent(fixture.ctx, fixture.eventID, joiner.ID, true, false, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get event with group")
	})

	t.Run("user game lookup fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationForEvent(fixture.ctx, fixture.eventID, joiner.ID, true, false, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get user game for registration")
	})

	t.Run("get registration fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 3, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationForEvent(fixture.ctx, fixture.eventID, joiner.ID, true, false, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get registration")
	})

	t.Run("create registration fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 4, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationForEvent(fixture.ctx, fixture.eventID, joiner.ID, true, false, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create registration")
	})

	t.Run("update registration fails", func(t *testing.T) {
		registerUserForEvent(t, fixture.ctx, fixture.tx, fixture.eventID, joiner.ID)
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 4, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationForEvent(fixture.ctx, fixture.eventID, joiner.ID, false, true, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "update registration")
	})
}

func TestUpsertRegistrationsForGroup_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)
	participant := createTestUser(t, fixture.ctx, fixture.s)
	createCompleteUserGameForRegistration(t, fixture.ctx, fixture.tx, fixture.s, participant.ID, fixture.gameID)
	items := []store.RegistrationUpsertItem{{EventID: fixture.eventID, CanSubstitute: true, CanLobbyHost: false}}

	t.Run("get event group fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationsForGroup(fixture.ctx, fixture.groupID, participant.ID, items, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get event group")
	})

	t.Run("get group events fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationsForGroup(fixture.ctx, fixture.groupID, participant.ID, items, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get group events")
	})

	t.Run("get registration fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 3, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationsForGroup(fixture.ctx, fixture.groupID, participant.ID, items, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get registration")
	})

	t.Run("create registration fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 4, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationsForGroup(fixture.ctx, fixture.groupID, participant.ID, items, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create registration")
	})

	t.Run("update registration fails", func(t *testing.T) {
		registerUserForEvent(t, fixture.ctx, fixture.tx, fixture.eventID, participant.ID)
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 4, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationsForGroup(fixture.ctx, fixture.groupID, participant.ID, items, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "update registration")
	})

	t.Run("delete unselected registration fails", func(t *testing.T) {
		event2 := insertEventInGroupFixture(t, fixture.ctx, fixture.tx, fixture.groupID, fixture.modeID, time.Now().UTC().Add(48*time.Hour))
		registerUserForEvent(t, fixture.ctx, fixture.tx, event2, participant.ID)
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnExecCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.UpsertRegistrationsForGroup(fixture.ctx, fixture.groupID, participant.ID, items, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete unselected registration")
	})
}

func TestDeleteRegistrationForEvent_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)
	joiner := createTestUser(t, fixture.ctx, fixture.s)
	registerUserForEvent(t, fixture.ctx, fixture.tx, fixture.eventID, joiner.ID)

	t.Run("get event fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.DeleteRegistrationForEvent(fixture.ctx, fixture.eventID, joiner.ID, joiner.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get event with group")
	})

	t.Run("get registration fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.DeleteRegistrationForEvent(fixture.ctx, fixture.eventID, joiner.ID, joiner.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get registration before delete")
	})

	t.Run("delete registration fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnExecCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		err := fs.DeleteRegistrationForEvent(fixture.ctx, fixture.eventID, joiner.ID, joiner.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete registration")
	})
}

func TestGetEventsForUser_QueryError(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	injectedErr := errors.New("injected database failure")
	faulty := &faultInjectTx{DBTX: tx, failOnQueryCall: 1, injectedErr: injectedErr}
	fs := store.NewPostgresStoreFromDBTXForTest(faulty)

	_, _, _, err := fs.GetEventsForUser(ctx, host.ID, true, false, nil, nil, "", "", "UTC")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "querying events for user")
}
