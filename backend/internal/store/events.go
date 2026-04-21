package store

import (
	"context"
	"fmt"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
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

func (s *PostgresStore) CreateEventGroupWithEvents(ctx context.Context, userID, gameModeID uuid.UUID, subMin int32, registrationOpen bool, region string, startTime time.Time, gamesToRun int32) (uuid.UUID, error) {
	var groupID uuid.UUID

	err := s.WithTx(ctx, func(tx Store) error {
		txStore, ok := tx.(*PostgresStore)
		if !ok {
			return fmt.Errorf("unexpected tx store type %T", tx)
		}

		gameMode, err := txStore.q.GetGameModeById(ctx, gameModeID)
		if err != nil {
			return fmt.Errorf("get game mode by id: %w", err)
		}

		group, err := txStore.q.CreateEventGroup(ctx, db.CreateEventGroupParams{
			OwnerID:          userID,
			GameModeID:       gameModeID,
			SubMin:           subMin,
			RegistrationOpen: registrationOpen,
			Region:           region,
		})
		if err != nil {
			return fmt.Errorf("create event group: %w", err)
		}
		groupID = group.ID

		nextStart := startTime
		for i := int32(0); i < gamesToRun; i++ {
			eventStart := nextStart
			err = txStore.q.CreateEvent(ctx, db.CreateEventParams{
				GroupID:   &groupID,
				StartTime: eventStart,
			})
			if err != nil {
				return fmt.Errorf("create event #%d: %w", i+1, err)
			}

			nextStart = nextStart.Add(time.Duration(gameMode.Duration) * time.Minute)
		}

		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	return groupID, nil
}
