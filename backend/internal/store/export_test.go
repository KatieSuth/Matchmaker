package store

import (
	"context"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
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
var ResolveSwapDestinationForTest = func(source SwapPlacementForTest, canSubstitute bool) SwapPlacementForTest {
	return swapPlacementToTest(resolveSwapDestination(swapPlacementFromTest(source), canSubstitute))
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
