package store

import (
	"context"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
)

// Expose unexported cursor helpers for black-box tests in store_test.
type DashboardCursorForTest = dashboardCursor

var EncodeDashboardCursorForTest = encodeDashboardCursor
var DecodeDashboardCursorForTest = decodeDashboardCursor
var IntPtrToInt32PtrForTest = intPtrToInt32Ptr
var PersistTeamPlansForTest = func(s *PostgresStore, ctx context.Context, plans []matchmaking.GamePlan) error {
	return s.persistTeamPlans(ctx, plans)
}
var GetEventGroupByIdForTest = func(s *PostgresStore, ctx context.Context, id uuid.UUID) (db.EventGroup, error) {
	return s.q.GetEventGroupById(ctx, id)
}
var PlanTeamsForGroupForTest = func(s *PostgresStore, ctx context.Context, group db.EventGroup, settings matchmaking.Settings) ([]matchmaking.GamePlan, error) {
	return s.planTeamsForGroup(ctx, group, settings)
}
var BuildGroupRegistrationCountsForTest = func(s *PostgresStore, ctx context.Context, groupID uuid.UUID) (map[uuid.UUID]int, error) {
	return s.buildGroupRegistrationCounts(ctx, groupID)
}
var MapRegistrationsToPlayersForTest = mapRegistrationsToPlayers
var BuildMatchmakingConfigForTest = func(group db.EventGroup) matchmaking.Config {
	return buildMatchmakingConfig(db.Event{}, group, modelGameModeInfo{Name: "Mode", TeamSize: 2}, 25, 0)
}
var SetPlayerAbsenceForEventAtForTest = func(s *PostgresStore, ctx context.Context, eventID, ownerID, userID uuid.UUID, absent bool, now time.Time) error {
	return s.setPlayerAbsenceForEventAt(ctx, eventID, ownerID, userID, absent, now)
}
var NoShowWindowForTest = noShowWindow
var ExcessiveNoShowsForDetailForTest = func(s *PostgresStore, ctx context.Context, hostID uuid.UUID, threshold int, events []model.EventGroupEvent, now time.Time) (map[uuid.UUID]int, error) {
	return s.excessiveNoShowsForDetail(ctx, hostID, threshold, events, now)
}
var BuildGroupNoShowCountsForTest = func(s *PostgresStore, ctx context.Context, group db.EventGroup, regCounts map[uuid.UUID]int, now time.Time) (map[uuid.UUID]int, error) {
	return s.buildGroupNoShowCounts(ctx, group, regCounts, now)
}
var VisibleAbsentUserIDsForTest = visibleAbsentUserIDs

var ResolveAvgRankIDForTest = func(s *PostgresStore, ctx context.Context, gameID uuid.UUID, currentOrder, peakOrder int32) (uuid.UUID, int32, error) {
	return s.resolveAvgRankID(ctx, gameID, currentOrder, peakOrder)
}

var EnsureAvgRanksForMatchmakingForTest = func(s *PostgresStore, ctx context.Context, rows []db.GetMatchmakingRegistrationsForEventRow) ([]db.GetMatchmakingRegistrationsForEventRow, error) {
	return s.ensureAvgRanksForMatchmaking(ctx, rows)
}

// SwapPlacementForTest mirrors swapPlacement for black-box unit tests.
type SwapPlacementForTest struct {
	Placed     bool
	LobbyID    uuid.UUID
	TeamNumber *int32
}

// swapPlacementFromTest converts exported test fixtures into internal swap placements.
func swapPlacementFromTest(p SwapPlacementForTest) swapPlacement {
	return swapPlacement{
		placed:     p.Placed,
		lobbyID:    p.LobbyID,
		teamNumber: p.TeamNumber,
	}
}

// swapPlacementToTest converts internal swap placements into exported test fixtures.
func swapPlacementToTest(p swapPlacement) SwapPlacementForTest {
	return SwapPlacementForTest{
		Placed:     p.placed,
		LobbyID:    p.lobbyID,
		TeamNumber: p.teamNumber,
	}
}

// SameTeamPlacementForTest exposes sameTeamPlacement for black-box unit tests.
var SameTeamPlacementForTest = func(a, b SwapPlacementForTest) bool {
	return sameTeamPlacement(swapPlacementFromTest(a), swapPlacementFromTest(b))
}

// SameSubPoolPlacementForTest exposes sameSubPoolPlacement for black-box unit tests.
var SameSubPoolPlacementForTest = func(a, b SwapPlacementForTest) bool {
	return sameSubPoolPlacement(swapPlacementFromTest(a), swapPlacementFromTest(b))
}

// ResolveSwapDestinationForTest exposes resolveSwapDestination for black-box unit tests.
var ResolveSwapDestinationForTest = func(ownPrevious, counterpart SwapPlacementForTest, canSubstitute bool) SwapPlacementForTest {
	return swapPlacementToTest(resolveSwapDestination(
		swapPlacementFromTest(ownPrevious),
		swapPlacementFromTest(counterpart),
		canSubstitute,
	))
}

// ValidateSubMinimumAfterSwapForTest exposes validateSubMinimumAfterSwap for black-box unit tests.
var ValidateSubMinimumAfterSwapForTest = func(
	before map[uuid.UUID]SwapPlacementForTest,
	userA, userB uuid.UUID,
	newA, newB SwapPlacementForTest,
	lobbyIDs []uuid.UUID,
	subMin int,
) error {
	internal := make(map[uuid.UUID]swapPlacement, len(before))
	for id, placement := range before {
		internal[id] = swapPlacementFromTest(placement)
	}
	lobbies := make([]db.GetLobbiesForEventRow, len(lobbyIDs))
	for i, id := range lobbyIDs {
		lobbies[i] = db.GetLobbiesForEventRow{ID: id}
	}
	return validateSubMinimumAfterSwap(
		internal,
		userA,
		userB,
		swapPlacementFromTest(newA),
		swapPlacementFromTest(newB),
		lobbies,
		subMin,
	)
}

// RecomputeLobbyAfterSwapForTest exposes recomputeLobbyAfterSwap for integration tests.
func NewPostgresStoreFromDBTXForTest(dbtx db.DBTX) *PostgresStore {
	return &PostgresStore{q: db.New(dbtx)}
}

var RecomputeLobbyAfterSwapForTest = func(
	s *PostgresStore,
	ctx context.Context,
	lobbyID uuid.UUID,
	settings matchmaking.Settings,
	gameModeID uuid.UUID,
) error {
	return s.recomputeLobbyAfterSwap(ctx, lobbyID, settings, gameModeID)
}
