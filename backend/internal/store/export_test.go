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
