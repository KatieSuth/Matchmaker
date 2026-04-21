package store

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
)

const dashboardEventsPageSize = 20

var (
	ErrInvalidGameID   = errors.New("invalid game_id")
	ErrInvalidCursor   = errors.New("invalid cursor")
	ErrInvalidTimezone = errors.New("invalid timezone")
)

type dashboardCursor struct {
	EventDate time.Time `json:"event_date"`
	ID        uuid.UUID `json:"id"`
}

func (s *PostgresStore) GetEventsForUser(ctx context.Context, userID uuid.UUID, hosting, past bool, from, to *time.Time, gameId, cursor, timezone string) ([]model.DashboardEvent, bool, string, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, false, "", fmt.Errorf("%w: %v", ErrInvalidTimezone, err)
	}

	nowInLocation := time.Now().In(location)
	dayStartInLocation := time.Date(nowInLocation.Year(), nowInLocation.Month(), nowInLocation.Day(), 0, 0, 0, 0, location)
	boundaryTime := dayStartInLocation.UTC()

	hasFrom := from != nil
	hasTo := to != nil
	applyPastFilter := !hasFrom && !hasTo

	var fromTime time.Time
	if hasFrom {
		fromTime = from.UTC()
	}

	var toTime time.Time
	if hasTo {
		toTime = to.AddDate(0, 0, 1).UTC()
	}

	hasGameID := gameId != ""
	gameUUID := uuid.Nil
	if hasGameID {
		parsedGameID, err := uuid.Parse(gameId)
		if err != nil {
			return nil, false, "", fmt.Errorf("%w: %v", ErrInvalidGameID, err)
		}
		gameUUID = parsedGameID
	}

	hasCursor := cursor != ""
	cursorTime := time.Time{}
	cursorID := uuid.Nil
	if hasCursor {
		parsedCursor, err := decodeDashboardCursor(cursor)
		if err != nil {
			return nil, false, "", fmt.Errorf("%w: %v", ErrInvalidCursor, err)
		}
		cursorTime = parsedCursor.EventDate.UTC()
		cursorID = parsedCursor.ID
	}

	rows, err := s.q.GetEventsForUser(ctx, db.GetEventsForUserParams{
		Hosting:         hosting,
		UserID:          userID,
		ApplyPastFilter: applyPastFilter,
		Past:            past,
		BoundaryTime:    boundaryTime,
		HasFrom:         hasFrom,
		FromTime:        fromTime,
		HasTo:           hasTo,
		ToTime:          toTime,
		HasGameID:       hasGameID,
		GameID:          gameUUID,
		HasCursor:       hasCursor,
		CursorTime:      cursorTime,
		CursorID:        cursorID,
		LimitCount:      dashboardEventsPageSize + 1,
	})
	if err != nil {
		return nil, false, "", fmt.Errorf("querying events for user: %w", err)
	}

	hasMore := len(rows) > dashboardEventsPageSize
	if hasMore {
		rows = rows[:dashboardEventsPageSize]
	}

	events := make([]model.DashboardEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, model.DashboardEvent{
			ID:               row.ID,
			GameName:         row.GameName,
			GameMode:         row.GameMode,
			EventDate:        row.EventDate,
			HostID:           row.HostID,
			HostName:         row.HostName,
			RegisteredCount:  int(row.RegisteredCount),
			RegistrationOpen: row.RegistrationOpen,
		})
	}

	if !hasMore || len(events) == 0 {
		return events, false, "", nil
	}

	nextCursor, err := encodeDashboardCursor(dashboardCursor{
		EventDate: events[len(events)-1].EventDate.UTC(),
		ID:        events[len(events)-1].ID,
	})
	if err != nil {
		return nil, false, "", fmt.Errorf("encoding dashboard cursor: %w", err)
	}

	return events, true, nextCursor, nil
}

func encodeDashboardCursor(cursor dashboardCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeDashboardCursor(cursor string) (dashboardCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return dashboardCursor{}, err
	}

	var parsed dashboardCursor
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return dashboardCursor{}, err
	}

	if parsed.ID == uuid.Nil || parsed.EventDate.IsZero() {
		return dashboardCursor{}, errors.New("cursor missing required fields")
	}

	return parsed, nil
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
