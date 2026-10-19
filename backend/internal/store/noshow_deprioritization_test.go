package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// insertNoShow records a no-show row directly, bypassing the host/team checks.
func insertNoShow(t *testing.T, ctx context.Context, tx db.DBTX, eventID, userID, hostID uuid.UUID) {
	t.Helper()
	_, err := tx.Exec(ctx, `INSERT INTO no_shows (event_id, user_id, host_id) VALUES ($1, $2, $3)`, eventID, userID, hostID)
	require.NoError(t, err)
}

// enableDeprioritization turns the setting on for a group and returns the refreshed row.
func enableDeprioritization(t *testing.T, f eventGroupFixture, threshold int32) db.EventGroup {
	t.Helper()
	_, err := f.tx.Exec(f.ctx, `UPDATE event_groups SET deprioritize_noshows = TRUE, max_noshows = $2 WHERE id = $1`, f.groupID, threshold)
	require.NoError(t, err)
	group, err := store.GetEventGroupByIdForTest(f.s, f.ctx, f.groupID)
	require.NoError(t, err)
	return group
}

func TestBuildGroupNoShowCounts_CountsHostGamesInWindowByStartTime(t *testing.T) {
	f := setupEventGroupFixture(t)
	now := time.Now().UTC()
	otherHost := createTestUser(t, f.ctx, f.s)
	player := createTestUser(t, f.ctx, f.s)
	other := createTestUser(t, f.ctx, f.s)

	_, recent := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, now.AddDate(0, -1, 0))
	_, recent2 := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, now.AddDate(0, -5, -20))
	_, stale := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, now.AddDate(0, -7, 0))
	_, future := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, now.Add(48*time.Hour))
	_, otherHostGame := insertEventFixture(t, f.ctx, f.tx, otherHost.ID, f.modeID, now.AddDate(0, -1, 0))

	insertNoShow(t, f.ctx, f.tx, recent, player.ID, f.hostID)
	insertNoShow(t, f.ctx, f.tx, recent2, player.ID, f.hostID)
	insertNoShow(t, f.ctx, f.tx, stale, player.ID, f.hostID)
	insertNoShow(t, f.ctx, f.tx, future, player.ID, f.hostID)
	insertNoShow(t, f.ctx, f.tx, otherHostGame, player.ID, otherHost.ID)
	insertNoShow(t, f.ctx, f.tx, recent, other.ID, f.hostID)

	group := enableDeprioritization(t, f, 2)
	counts, err := store.BuildGroupNoShowCountsForTest(f.s, f.ctx, group, map[uuid.UUID]int{player.ID: 1}, now)
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]int{player.ID: 2}, counts, "only this host's games within 6 months; registrants only")
}

func TestBuildGroupNoShowCounts_DisabledSkipsQuery(t *testing.T) {
	f := setupEventGroupFixture(t)
	_, err := f.tx.Exec(f.ctx, `UPDATE event_groups SET max_noshows = 1 WHERE id = $1`, f.groupID)
	require.NoError(t, err)
	group, err := store.GetEventGroupByIdForTest(f.s, f.ctx, f.groupID)
	require.NoError(t, err)

	faulty := &faultInjectTx{DBTX: f.tx, failOn: "CountRecentNoShowsForHost", injectedErr: errors.New("must not query")}
	counts, err := store.BuildGroupNoShowCountsForTest(store.NewPostgresStoreFromDBTXForTest(faulty), f.ctx, group, map[uuid.UUID]int{uuid.New(): 1}, time.Now())
	require.NoError(t, err)
	assert.Nil(t, counts)
	assert.Zero(t, faulty.totalCalls)
}

func TestBuildMatchmakingConfig_ThresholdOnlyWhenEnabled(t *testing.T) {
	assert.Equal(t, 0, store.BuildMatchmakingConfigForTest(db.EventGroup{DeprioritizeNoshows: false, MaxNoshows: 5}).NoShowThreshold)
	assert.Equal(t, 5, store.BuildMatchmakingConfigForTest(db.EventGroup{DeprioritizeNoshows: true, MaxNoshows: 5}).NoShowThreshold)
}

// registerFivePlayers registers five equally ranked non-subs for a 2v2 game and returns their IDs.
func registerFivePlayers(t *testing.T, f eventGroupFixture) []uuid.UUID {
	t.Helper()
	ids := make([]uuid.UUID, 0, 5)
	for i := 0; i < 5; i++ {
		u := createTestUser(t, f.ctx, f.s)
		registerPlayerForEventWithProfile(t, f.ctx, f.tx, f.s, f.eventID, u.ID, f.gameID, false, false)
		ids = append(ids, u.ID)
	}
	return ids
}

// rosteredInPlans reports which users hold a roster seat in any planned lobby.
func rosteredInPlans(plans []matchmaking.GamePlan) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for _, plan := range plans {
		for _, lobby := range plan.Lobbies {
			for _, p := range lobby.Roster {
				out[p.UserID] = true
			}
		}
	}
	return out
}

func TestPlanTeamsForGroup_DeprioritizesRecentNoShows(t *testing.T) {
	f := setupEventGroupFixture(t)
	ids := registerFivePlayers(t, f)
	excessive := ids[0]
	_, past1 := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, time.Now().UTC().AddDate(0, -1, 0))
	_, past2 := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, time.Now().UTC().AddDate(0, -2, 0))
	insertNoShow(t, f.ctx, f.tx, past1, excessive, f.hostID)
	insertNoShow(t, f.ctx, f.tx, past2, excessive, f.hostID)

	group := enableDeprioritization(t, f, 2)
	plans, err := store.PlanTeamsForGroupForTest(f.s, f.ctx, group, defaultMatchmakingSettings())
	require.NoError(t, err)
	roster := rosteredInPlans(plans)
	assert.Len(t, roster, 4)
	assert.False(t, roster[excessive], "two recent no-shows at threshold 2 sits out when four others can play")
}

func TestPlanTeamsForGroup_DisabledIgnoresNoShows(t *testing.T) {
	f := setupEventGroupFixture(t)
	ids := registerFivePlayers(t, f)
	_, past := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, time.Now().UTC().AddDate(0, -1, 0))
	for _, id := range ids[:4] {
		insertNoShow(t, f.ctx, f.tx, past, id, f.hostID)
	}
	// A stored threshold with the toggle off must have no effect.
	_, err := f.tx.Exec(f.ctx, `UPDATE event_groups SET max_noshows = 1 WHERE id = $1`, f.groupID)
	require.NoError(t, err)
	group, err := store.GetEventGroupByIdForTest(f.s, f.ctx, f.groupID)
	require.NoError(t, err)

	plans, err := store.PlanTeamsForGroupForTest(f.s, f.ctx, group, defaultMatchmakingSettings())
	require.NoError(t, err)
	for _, plan := range plans {
		for _, lobby := range plan.Lobbies {
			for _, p := range lobby.Roster {
				assert.Zero(t, p.NoShowCount)
				assert.False(t, p.Deprioritized)
			}
		}
	}
}

func TestPlanTeamsForGroup_NoShowCountQueryFails(t *testing.T) {
	f := setupEventGroupFixture(t)
	registerFivePlayers(t, f)
	group := enableDeprioritization(t, f, 2)

	// Query order: events (plan), events + registrations (counts), then the no-show count.
	faulty := &faultInjectTx{DBTX: f.tx, failOn: "CountRecentNoShowsForHost", injectedErr: errors.New("injected database failure")}
	_, err := store.PlanTeamsForGroupForTest(store.NewPostgresStoreFromDBTXForTest(faulty), f.ctx, group, defaultMatchmakingSettings())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count recent no-shows for host")
}

func TestEventGroupSettings_PersistNoShowPolicy(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)
	start := time.Now().UTC().Add(24 * time.Hour)

	groupID, err := s.CreateEventGroupWithEvents(ctx, host.ID, modeID, 0, true, "AMER", "balanced", "", start, 1, nil, store.NoShowPolicy{Enabled: true, Threshold: 3})
	require.NoError(t, err)
	detail, err := s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	assert.True(t, detail.DeprioritizeNoshows)
	assert.Equal(t, 3, detail.MaxNoshows)
	updates := patchEventUpdates(detail.Events[0].ID, modeID, start)

	// nil keeps the stored setting.
	require.NoError(t, s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, "", updates, nil, nil))
	detail, err = s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	assert.True(t, detail.DeprioritizeNoshows)
	assert.Equal(t, 3, detail.MaxNoshows)

	// Disabling keeps the threshold so re-enabling restores it; the sent threshold is ignored.
	require.NoError(t, s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, "", updates, nil, &store.NoShowPolicy{Enabled: false, Threshold: -4}))
	detail, err = s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	assert.False(t, detail.DeprioritizeNoshows)
	assert.Equal(t, 3, detail.MaxNoshows)

	require.NoError(t, s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, "", updates, nil, &store.NoShowPolicy{Enabled: true, Threshold: 5}))
	detail, err = s.GetEventGroupDetail(ctx, groupID, host.ID)
	require.NoError(t, err)
	assert.True(t, detail.DeprioritizeNoshows)
	assert.Equal(t, 5, detail.MaxNoshows)

	err = s.UpdateEventGroupSettings(ctx, groupID, host.ID, "AMER", 0, "balanced", true, "", updates, nil, &store.NoShowPolicy{Enabled: true, Threshold: 0})
	assert.ErrorIs(t, err, store.ErrInvalidNoShowThreshold)
}

func TestCreateEventGroupWithEvents_RejectsEnabledZeroThreshold(t *testing.T) {
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	host := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	modeID := insertSmallTeamMode(t, ctx, tx, games[0].ID)

	_, err = s.CreateEventGroupWithEvents(ctx, host.ID, modeID, 0, true, "AMER", "balanced", "", time.Now().UTC().Add(time.Hour), 1, nil, store.NoShowPolicy{Enabled: true})
	assert.ErrorIs(t, err, store.ErrInvalidNoShowThreshold)
}

func TestEventGroups_CheckRejectsEnabledWithoutThreshold(t *testing.T) {
	f := setupEventGroupFixture(t)
	_, err := f.tx.Exec(f.ctx, `UPDATE event_groups SET deprioritize_noshows = TRUE, max_noshows = 0 WHERE id = $1`, f.groupID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event_groups_deprioritize_noshows_threshold")
}

// excessiveFixture registers three players for the fixture game: one at the threshold, one
// above it, and one below, each with recent no-shows at the host.
func excessiveFixture(t *testing.T) (f eventGroupFixture, atLimit, above, below uuid.UUID) {
	t.Helper()
	f = setupEventGroupFixture(t)
	ids := make([]uuid.UUID, 3)
	for i := range ids {
		u := createTestUser(t, f.ctx, f.s)
		registerPlayerForEventWithProfile(t, f.ctx, f.tx, f.s, f.eventID, u.ID, f.gameID, false, false)
		ids[i] = u.ID
	}
	var past []uuid.UUID
	for i := 0; i < 3; i++ {
		_, ev := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, time.Now().UTC().AddDate(0, 0, -7*(i+1)))
		past = append(past, ev)
	}
	for _, ev := range past[:2] {
		insertNoShow(t, f.ctx, f.tx, ev, ids[0], f.hostID)
	}
	for _, ev := range past {
		insertNoShow(t, f.ctx, f.tx, ev, ids[1], f.hostID)
	}
	insertNoShow(t, f.ctx, f.tx, past[0], ids[2], f.hostID)
	return f, ids[0], ids[1], ids[2]
}

func TestGetEventGroupDetail_ExcessiveNoShowsHostOnly(t *testing.T) {
	f, atLimit, above, below := excessiveFixture(t)
	enableDeprioritization(t, f, 2)

	detail, err := f.s.GetEventGroupDetail(f.ctx, f.groupID, f.hostID)
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]int{atLimit: 2, above: 3}, detail.ExcessiveNoShows, "below-threshold players are not flagged")
	assert.NotContains(t, detail.ExcessiveNoShows, below)

	for _, viewer := range []uuid.UUID{atLimit, createTestUser(t, f.ctx, f.s).ID} {
		detail, err = f.s.GetEventGroupDetail(f.ctx, f.groupID, viewer)
		require.NoError(t, err)
		assert.NotNil(t, detail.ExcessiveNoShows)
		assert.Empty(t, detail.ExcessiveNoShows, "only the host sees no-show counts")
	}
}

func TestGetEventGroupDetail_ExcessiveNoShowsEmptyWhenDeprioritizationOff(t *testing.T) {
	f, _, _, _ := excessiveFixture(t)
	_, err := f.tx.Exec(f.ctx, `UPDATE event_groups SET max_noshows = 1 WHERE id = $1`, f.groupID)
	require.NoError(t, err)

	detail, err := f.s.GetEventGroupDetail(f.ctx, f.groupID, f.hostID)
	require.NoError(t, err)
	assert.NotNil(t, detail.ExcessiveNoShows)
	assert.Empty(t, detail.ExcessiveNoShows)
}

func TestGetEventGroupDetail_ExcessiveNoShowsNoRegistrants(t *testing.T) {
	f := setupEventGroupFixture(t)
	enableDeprioritization(t, f, 2)

	detail, err := f.s.GetEventGroupDetail(f.ctx, f.groupID, f.hostID)
	require.NoError(t, err)
	assert.Empty(t, detail.ExcessiveNoShows)
}

func TestExcessiveNoShowsForDetail_NonPositiveThresholdFlagsNobody(t *testing.T) {
	f, _, _, _ := excessiveFixture(t)
	detail, err := f.s.GetEventGroupDetail(f.ctx, f.groupID, f.hostID)
	require.NoError(t, err)

	got, err := store.ExcessiveNoShowsForDetailForTest(f.s, f.ctx, f.hostID, 0, detail.Events, time.Now())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetEventGroupDetail_ExcessiveNoShowsQueryFails(t *testing.T) {
	f, _, _, _ := excessiveFixture(t)
	enableDeprioritization(t, f, 2)

	// Query order: events summary, registrations, per-game no-show list, Discord guilds,
	// viewer no-show dates, then the count.
	faulty := &faultInjectTx{DBTX: f.tx, failOn: "CountRecentNoShowsForHost", injectedErr: errors.New("injected database failure")}
	_, err := store.NewPostgresStoreFromDBTXForTest(faulty).GetEventGroupDetail(f.ctx, f.groupID, f.hostID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count recent no-shows for host")
}

func TestGetEventGroupDetail_ViewerNoShowDatesOwnRecordOnly(t *testing.T) {
	f := setupEventGroupFixture(t)
	now := time.Now().UTC().Truncate(time.Second)
	otherHost := createTestUser(t, f.ctx, f.s)
	viewer := createTestUser(t, f.ctx, f.s)
	bystander := createTestUser(t, f.ctx, f.s)

	older := now.AddDate(0, -2, 0)
	newer := now.AddDate(0, 0, -7)
	_, evNewer := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, newer)
	_, evOlder := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, older)
	_, evStale := insertEventFixture(t, f.ctx, f.tx, f.hostID, f.modeID, now.AddDate(0, -7, 0))
	_, evOtherHost := insertEventFixture(t, f.ctx, f.tx, otherHost.ID, f.modeID, newer)
	insertNoShow(t, f.ctx, f.tx, evNewer, viewer.ID, f.hostID)
	insertNoShow(t, f.ctx, f.tx, evOlder, viewer.ID, f.hostID)
	insertNoShow(t, f.ctx, f.tx, evStale, viewer.ID, f.hostID)
	insertNoShow(t, f.ctx, f.tx, evOtherHost, viewer.ID, otherHost.ID)
	insertNoShow(t, f.ctx, f.tx, evNewer, bystander.ID, f.hostID)

	// Deprioritization off: the record is not sent at all.
	detail, err := f.s.GetEventGroupDetail(f.ctx, f.groupID, viewer.ID)
	require.NoError(t, err)
	assert.NotNil(t, detail.ViewerNoShowDates)
	assert.Empty(t, detail.ViewerNoShowDates)

	enableDeprioritization(t, f, 3)
	detail, err = f.s.GetEventGroupDetail(f.ctx, f.groupID, viewer.ID)
	require.NoError(t, err)
	require.Len(t, detail.ViewerNoShowDates, 2, "this host's games in the last 6 months only")
	assert.True(t, older.Equal(detail.ViewerNoShowDates[0]), "oldest first")
	assert.True(t, newer.Equal(detail.ViewerNoShowDates[1]))

	detail, err = f.s.GetEventGroupDetail(f.ctx, f.groupID, f.hostID)
	require.NoError(t, err)
	assert.NotNil(t, detail.ViewerNoShowDates)
	assert.Empty(t, detail.ViewerNoShowDates, "the host is never shown another player's dates")
}

func TestGetEventGroupDetail_ViewerNoShowDatesQueryFails(t *testing.T) {
	f := setupEventGroupFixture(t)
	enableDeprioritization(t, f, 2)
	// Query order: events summary, registrations, per-game no-show list, Discord guilds, then viewer dates.
	faulty := &faultInjectTx{DBTX: f.tx, failOn: "ListRecentNoShowDatesForHostUser", injectedErr: errors.New("injected database failure")}
	_, err := store.NewPostgresStoreFromDBTXForTest(faulty).GetEventGroupDetail(f.ctx, f.groupID, f.hostID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list viewer no-show dates")
}

func TestNoShowWindow(t *testing.T) {
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	since, until := store.NoShowWindowForTest(now)
	assert.Equal(t, time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC), since)
	assert.Equal(t, now, until)
}
