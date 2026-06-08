package store_test

import (
	"errors"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGroupRegistrationCounts_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)
	joiner := createTestUser(t, fixture.ctx, fixture.s)
	registerUserForEvent(t, fixture.ctx, fixture.tx, fixture.eventID, joiner.ID)

	t.Run("get events fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := store.BuildGroupRegistrationCountsForTest(fs, fixture.ctx, fixture.groupID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get events by group")
	})

	t.Run("get registrations fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := store.BuildGroupRegistrationCountsForTest(fs, fixture.ctx, fixture.groupID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get registrations for event")
	})
}

func TestBuildGroupRegistrationCounts_Success(t *testing.T) {
	fixture := setupEventGroupFixture(t)
	joiner := createTestUser(t, fixture.ctx, fixture.s)
	registerUserForEvent(t, fixture.ctx, fixture.tx, fixture.eventID, joiner.ID)
	event2 := insertEventInGroupFixture(t, fixture.ctx, fixture.tx, fixture.groupID, fixture.modeID, time.Now().UTC().Add(48*time.Hour))
	registerUserForEvent(t, fixture.ctx, fixture.tx, event2, joiner.ID)

	counts, err := store.BuildGroupRegistrationCountsForTest(fixture.s, fixture.ctx, fixture.groupID)
	require.NoError(t, err)
	assert.Equal(t, 2, counts[joiner.ID])
}

func TestPlanTeamsForGroup_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	fixture := setupEventGroupFixture(t)
	for i := 0; i < 4; i++ {
		u := createTestUser(t, fixture.ctx, fixture.s)
		registerPlayerForEventWithProfile(t, fixture.ctx, fixture.tx, fixture.s, fixture.eventID, u.ID, fixture.gameID, false, false)
	}
	group, err := store.GetEventGroupByIdForTest(fixture.s, fixture.ctx, fixture.groupID)
	require.NoError(t, err)

	t.Run("get events fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := store.PlanTeamsForGroupForTest(fs, fixture.ctx, group, defaultMatchmakingSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get events by group")
	})

	t.Run("get registrations for counts fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryCall: 3, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := store.PlanTeamsForGroupForTest(fs, fixture.ctx, group, defaultMatchmakingSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get registrations for event")
	})

	t.Run("get matchmaking registrations fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryCall: 4, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := store.PlanTeamsForGroupForTest(fs, fixture.ctx, group, defaultMatchmakingSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get matchmaking registrations for event")
	})

	t.Run("get game mode fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 1, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := store.PlanTeamsForGroupForTest(fs, fixture.ctx, group, defaultMatchmakingSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get game mode for event")
	})

	t.Run("get max rank order fails", func(t *testing.T) {
		faulty := &faultInjectTx{DBTX: fixture.tx, failOnQueryRowCall: 2, injectedErr: injectedErr}
		fs := store.NewPostgresStoreFromDBTXForTest(faulty)

		_, err := store.PlanTeamsForGroupForTest(fs, fixture.ctx, group, defaultMatchmakingSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "get max rank order for game")
	})
}

func TestPlanTeamsForGroup_SkipsEmptyEvent(t *testing.T) {
	fixture := setupEventGroupFixture(t)
	insertEventInGroupFixture(t, fixture.ctx, fixture.tx, fixture.groupID, fixture.modeID, time.Now().UTC().Add(48*time.Hour))

	for i := 0; i < 4; i++ {
		u := createTestUser(t, fixture.ctx, fixture.s)
		registerPlayerForEventWithProfile(t, fixture.ctx, fixture.tx, fixture.s, fixture.eventID, u.ID, fixture.gameID, false, false)
	}

	group, err := store.GetEventGroupByIdForTest(fixture.s, fixture.ctx, fixture.groupID)
	require.NoError(t, err)

	plans, err := store.PlanTeamsForGroupForTest(fixture.s, fixture.ctx, group, defaultMatchmakingSettings())
	require.NoError(t, err)
	require.Len(t, plans, 1)
	assert.Equal(t, fixture.eventID, plans[0].EventID)
}

func TestPlanTeamsForGroup_InsufficientPlayers(t *testing.T) {
	fixture := setupEventGroupFixture(t)
	u := createTestUser(t, fixture.ctx, fixture.s)
	registerPlayerForEventWithProfile(t, fixture.ctx, fixture.tx, fixture.s, fixture.eventID, u.ID, fixture.gameID, false, false)

	group, err := store.GetEventGroupByIdForTest(fixture.s, fixture.ctx, fixture.groupID)
	require.NoError(t, err)

	_, err = store.PlanTeamsForGroupForTest(fixture.s, fixture.ctx, group, defaultMatchmakingSettings())
	require.Error(t, err)
	var teamErr *store.TeamCreationError
	require.ErrorAs(t, err, &teamErr)
	assert.ErrorIs(t, teamErr, store.ErrInsufficientPlayers)
}
