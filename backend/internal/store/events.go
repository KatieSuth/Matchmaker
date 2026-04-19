package store

import (
	"context"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
)

func (s *PostgresStore) GetEventsForUser(ctx context.Context, userID uuid.UUID, hosting, past bool, from, to *time.Time, gameId, cursor string) ([]model.DashboardEvent, bool, string, error) {

	// TODO: Make the query for this.
	// There's no data in the system yet and no way to make the data
	// This is a placeholder for now.
	// Returns slice of dashboard events, bool to indicate there are no more records,
	//   empty string for nothing in the next cursor, and nil error
	return []model.DashboardEvent{}, false, "", nil

}
