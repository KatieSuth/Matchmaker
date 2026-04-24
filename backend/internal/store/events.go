package store

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const dashboardEventsPageSize = 20

var (
	ErrInvalidGameID         = errors.New("invalid game_id")
	ErrInvalidCursor         = errors.New("invalid cursor")
	ErrInvalidTimezone       = errors.New("invalid timezone")
	ErrForbidden             = errors.New("forbidden")
	ErrRegistrationClosed    = errors.New("registration is closed")
	ErrTeamsAlreadyCreated   = errors.New("teams already created")
	ErrTeamsNotCreated       = errors.New("teams not created")
	ErrRegistrationNotFound  = errors.New("registration not found")
	ErrInvalidSubMin         = errors.New("invalid sub_min")
	ErrOpenRegistrationTeams = errors.New("cannot open registration while teams exist")
	ErrEventGroupNotFound    = errors.New("event group not found")
	ErrEventNotFound         = errors.New("event not found")
	ErrGameModeNotFound      = errors.New("game mode not found")
)

// dashboardCursor is encoded into the opaque "next page" string for the user events list.
type dashboardCursor struct {
	EventDate time.Time `json:"event_date"`
	ID        uuid.UUID `json:"id"`
}

// GetEventsForUser returns dashboard rows with optional date filters, game filter, and
// keyset pagination. When from/to are both absent, applyPastFilter drives "today" vs
// past tabs using the caller’s local-time midnight boundary in UTC.
func (s *PostgresStore) GetEventsForUser(ctx context.Context, userID uuid.UUID, hosting, past bool, from, to *time.Time, gameId, cursor, timezone string) ([]model.DashboardEvent, bool, string, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, false, "", fmt.Errorf("%w: %v", ErrInvalidTimezone, err)
	}

	nowInLocation := time.Now().In(location)
	dayStartInLocation := time.Date(nowInLocation.Year(), nowInLocation.Month(), nowInLocation.Day(), 0, 0, 0, 0, location)
	// "Today" for the user in their zone, as an instant in UTC, for the default past/upcoming filter.
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
		// Inclusive end date: add one day and use a half-open [from, to) range in the query.
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

	// Fetch one extra row to know whether a next page exists without a separate count query.
	hasMore := len(rows) > dashboardEventsPageSize
	if hasMore {
		rows = rows[:dashboardEventsPageSize]
	}

	events := make([]model.DashboardEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, model.MapDbGetEventsForUserRowToDashboardEvent(row))
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
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrGameModeNotFound
			}
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

// GetEventGroupDetail loads a group header plus each scheduled game and, per game,
// the registration list the viewer may see (full detail for the host, limited otherwise).
func (s *PostgresStore) GetEventGroupDetail(ctx context.Context, groupID, viewerID uuid.UUID) (model.EventGroupDetail, error) {
	groupRow, err := s.q.GetEventGroupDetailById(ctx, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EventGroupDetail{}, ErrEventGroupNotFound
		}
		return model.EventGroupDetail{}, fmt.Errorf("get event group detail: %w", err)
	}

	summaries, err := s.q.GetGroupEventsSummary(ctx, db.GetGroupEventsSummaryParams{
		GroupID:  &groupID,
		ViewerID: viewerID,
	})
	if err != nil {
		return model.EventGroupDetail{}, fmt.Errorf("get group events summary: %w", err)
	}

	viewerIsHost := groupRow.OwnerID == viewerID
	events := make([]model.EventGroupEvent, 0, len(summaries))
	for _, summary := range summaries {
		regRows, err := s.q.GetRegistrationDataByEventId(ctx, db.GetRegistrationDataByEventIdParams{
			EventID:      summary.ID,
			ViewerIsHost: viewerIsHost,
		})
		if err != nil {
			return model.EventGroupDetail{}, fmt.Errorf("get registration data for event %s: %w", summary.ID.String(), err)
		}

		registrations := model.MapDbGetRegistrationDataByEventIdRowsToEventRegistrations(regRows)

		events = append(events, model.MapDbGetGroupEventsSummaryRowToEventGroupEvent(summary, registrations))
	}

	return model.MapDbGetEventGroupDetailByIdRowToEventGroupDetail(groupRow, events), nil
}

func (s *PostgresStore) UpdateEventGroupSettings(ctx context.Context, groupID, ownerID uuid.UUID, region string, subMin int32) error {
	group, err := s.q.GetEventGroupById(ctx, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEventGroupNotFound
		}
		return fmt.Errorf("get event group: %w", err)
	}
	if group.OwnerID != ownerID {
		return ErrForbidden
	}
	if strings.TrimSpace(region) == "" || subMin < 0 {
		return ErrInvalidSubMin
	}

	_, err = s.q.UpdateEventGroupSettings(ctx, db.UpdateEventGroupSettingsParams{
		ID:     groupID,
		Region: strings.TrimSpace(region),
		SubMin: subMin,
	})
	if err != nil {
		return fmt.Errorf("update event group settings: %w", err)
	}
	return nil
}

func (s *PostgresStore) DeleteEventGroup(ctx context.Context, groupID, ownerID uuid.UUID) error {
	group, err := s.q.GetEventGroupById(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get event group: %w", err)
	}
	if group.OwnerID != ownerID {
		return ErrForbidden
	}

	rowsDeleted, err := s.q.DeleteEventGroupById(ctx, groupID)
	if err != nil {
		return fmt.Errorf("delete event group: %w", err)
	}
	if rowsDeleted == 0 {
		return ErrEventGroupNotFound
	}
	return nil
}

func (s *PostgresStore) SetEventGroupRegistrationOpen(ctx context.Context, groupID, ownerID uuid.UUID, open bool) error {
	group, err := s.q.GetEventGroupById(ctx, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEventGroupNotFound
		}
		return fmt.Errorf("get event group: %w", err)
	}
	if group.OwnerID != ownerID {
		return ErrForbidden
	}
	if open {
		lobbyCount, err := s.q.CountLobbiesByGroupId(ctx, &groupID)
		if err != nil {
			return fmt.Errorf("count lobbies by group: %w", err)
		}
		if lobbyCount > 0 {
			return ErrOpenRegistrationTeams
		}
	}

	_, err = s.q.SetEventGroupRegistrationOpen(ctx, db.SetEventGroupRegistrationOpenParams{
		ID:               groupID,
		RegistrationOpen: open,
	})
	if err != nil {
		return fmt.Errorf("set registration_open: %w", err)
	}
	return nil
}

// CreateTeamsForGroup closes registration, then for each event with signups creates a lobby
// and assigns the first "can lobby host" registrant as host when available.
func (s *PostgresStore) CreateTeamsForGroup(ctx context.Context, groupID, ownerID uuid.UUID) error {
	group, err := s.q.GetEventGroupById(ctx, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEventGroupNotFound
		}
		return fmt.Errorf("get event group: %w", err)
	}
	if group.OwnerID != ownerID {
		return ErrForbidden
	}

	existingLobbies, err := s.q.CountLobbiesByGroupId(ctx, &groupID)
	if err != nil {
		return fmt.Errorf("count lobbies by group: %w", err)
	}
	if existingLobbies > 0 {
		return ErrTeamsAlreadyCreated
	}

	return s.WithTx(ctx, func(tx Store) error {
		txStore, ok := tx.(*PostgresStore)
		if !ok {
			return fmt.Errorf("unexpected tx store type %T", tx)
		}

		_, err := txStore.q.SetEventGroupRegistrationOpen(ctx, db.SetEventGroupRegistrationOpenParams{
			ID:               groupID,
			RegistrationOpen: false,
		})
		if err != nil {
			return fmt.Errorf("close registration for group: %w", err)
		}

		events, err := txStore.q.GetEventsByGroupId(ctx, &groupID)
		if err != nil {
			return fmt.Errorf("get events by group: %w", err)
		}

		for _, eventRow := range events {
			regs, err := txStore.q.GetRegistrationsForEvent(ctx, eventRow.ID)
			if err != nil {
				return fmt.Errorf("get registrations for event %s: %w", eventRow.ID.String(), err)
			}
			if len(regs) == 0 {
				continue
			}

			var lobbyHost *uuid.UUID
			for _, reg := range regs {
				if reg.CanLobbyHost {
					hostID := reg.UserID
					lobbyHost = &hostID
					break
				}
			}

			lobby, err := txStore.q.CreateLobby(ctx, db.CreateLobbyParams{
				EventID: &eventRow.ID,
				Host:    lobbyHost,
			})
			if err != nil {
				return fmt.Errorf("create lobby for event %s: %w", eventRow.ID.String(), err)
			}

			for _, reg := range regs {
				err = txStore.q.CreatePlayer(ctx, db.CreatePlayerParams{
					LobbyID: lobby.ID,
					UserID:  reg.UserID,
				})
				if err != nil {
					return fmt.Errorf("create player for event %s: %w", eventRow.ID.String(), err)
				}
			}
		}

		return nil
	})
}

func (s *PostgresStore) DeleteTeamsAndOpenRegistration(ctx context.Context, groupID, ownerID uuid.UUID) error {
	group, err := s.q.GetEventGroupById(ctx, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEventGroupNotFound
		}
		return fmt.Errorf("get event group: %w", err)
	}
	if group.OwnerID != ownerID {
		return ErrForbidden
	}

	existingLobbies, err := s.q.CountLobbiesByGroupId(ctx, &groupID)
	if err != nil {
		return fmt.Errorf("count lobbies by group: %w", err)
	}
	if existingLobbies == 0 {
		return ErrTeamsNotCreated
	}

	return s.WithTx(ctx, func(tx Store) error {
		txStore, ok := tx.(*PostgresStore)
		if !ok {
			return fmt.Errorf("unexpected tx store type %T", tx)
		}

		if err := txStore.q.DeletePlayersByGroupId(ctx, &groupID); err != nil {
			return fmt.Errorf("delete players by group: %w", err)
		}
		if err := txStore.q.DeleteLobbiesByGroupId(ctx, &groupID); err != nil {
			return fmt.Errorf("delete lobbies by group: %w", err)
		}
		if _, err := txStore.q.SetEventGroupRegistrationOpen(ctx, db.SetEventGroupRegistrationOpenParams{
			ID:               groupID,
			RegistrationOpen: true,
		}); err != nil {
			return fmt.Errorf("open registration for group: %w", err)
		}
		return nil
	})
}

func (s *PostgresStore) UpsertRegistrationForEvent(ctx context.Context, eventID, userID uuid.UUID, canSubstitute, canLobbyHost bool, duoRequest *string) error {
	eventRow, err := s.q.GetEventWithGroupById(ctx, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEventNotFound
		}
		return fmt.Errorf("get event with group: %w", err)
	}
	if !eventRow.RegistrationOpen {
		return ErrRegistrationClosed
	}

	_, err = s.q.GetRegistrationByEventAndUser(ctx, db.GetRegistrationByEventAndUserParams{
		EventID: eventID,
		UserID:  userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = s.q.CreateRegistration(ctx, db.CreateRegistrationParams{
				EventID:       eventID,
				UserID:        userID,
				CanSubstitute: canSubstitute,
				CanLobbyHost:  canLobbyHost,
				DuoRequest:    duoRequest,
			})
			if err != nil {
				return fmt.Errorf("create registration: %w", err)
			}
			return nil
		}
		return fmt.Errorf("get registration: %w", err)
	}

	_, err = s.q.UpdateRegistration(ctx, db.UpdateRegistrationParams{
		EventID:       eventID,
		UserID:        userID,
		CanSubstitute: canSubstitute,
		CanLobbyHost:  canLobbyHost,
		DuoRequest:    duoRequest,
	})
	if err != nil {
		return fmt.Errorf("update registration: %w", err)
	}
	return nil
}

func (s *PostgresStore) DeleteRegistrationForEvent(ctx context.Context, eventID, targetUserID, actorUserID uuid.UUID) error {
	eventRow, err := s.q.GetEventWithGroupById(ctx, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEventNotFound
		}
		return fmt.Errorf("get event with group: %w", err)
	}
	isHost := eventRow.OwnerID == actorUserID
	isSelf := targetUserID == actorUserID
	if !isHost && !isSelf {
		return ErrForbidden
	}
	if !isHost && !eventRow.RegistrationOpen {
		return ErrRegistrationClosed
	}

	_, err = s.q.GetRegistrationByEventAndUser(ctx, db.GetRegistrationByEventAndUserParams{
		EventID: eventID,
		UserID:  targetUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRegistrationNotFound
		}
		return fmt.Errorf("get registration before delete: %w", err)
	}

	if err := s.q.DeleteRegistration(ctx, db.DeleteRegistrationParams{
		EventID: eventID,
		UserID:  targetUserID,
	}); err != nil {
		return fmt.Errorf("delete registration: %w", err)
	}
	return nil
}
