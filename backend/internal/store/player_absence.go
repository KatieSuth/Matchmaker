package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrInvalidAbsenceChange is the sentinel for attendance validation failures surfaced to handlers.
var ErrInvalidAbsenceChange = errors.New("invalid absence change")

// AbsenceValidationError carries a client-facing message for attendance validation failures.
type AbsenceValidationError struct {
	Message string
}

func (e *AbsenceValidationError) Error() string {
	return e.Message
}

func (e *AbsenceValidationError) Unwrap() error {
	return ErrInvalidAbsenceChange
}

// SetPlayerAbsenceForEvent marks (absent=true) or clears (absent=false) a no-show for one locked-in game.
// Only the event-group owner may change attendance. Marking requires the game to have started and the
// player to be on a team; clearing is idempotent and skips both checks so a mark can still be removed
// after a swap moved the player off a team.
func (s *PostgresStore) SetPlayerAbsenceForEvent(
	ctx context.Context,
	eventID, ownerID, userID uuid.UUID,
	absent bool,
) error {
	return s.setPlayerAbsenceForEventAt(ctx, eventID, ownerID, userID, absent, time.Now())
}

// setPlayerAbsenceForEventAt is SetPlayerAbsenceForEvent with an injectable clock for tests.
func (s *PostgresStore) setPlayerAbsenceForEventAt(
	ctx context.Context,
	eventID, ownerID, userID uuid.UUID,
	absent bool,
	now time.Time,
) error {
	meta, err := s.q.GetEventGroupMetaByEventId(ctx, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEventNotFound
		}
		return fmt.Errorf("get event group meta: %w", err)
	}
	if meta.OwnerID != ownerID {
		return ErrForbidden
	}

	lobbyCount, err := s.q.CountLobbiesForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("count lobbies for event: %w", err)
	}
	if lobbyCount == 0 {
		return ErrTeamsNotCreated
	}

	if !absent {
		if err := s.q.DeleteNoShow(ctx, db.DeleteNoShowParams{EventID: eventID, UserID: userID}); err != nil {
			return fmt.Errorf("delete no-show: %w", err)
		}
		return nil
	}

	// A no-show can only happen once the game is under way.
	if meta.StartTime.After(now) {
		return &AbsenceValidationError{Message: "Players can be marked absent once the game has started"}
	}

	placements, err := s.q.GetPlayerPlacementsForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("get player placements: %w", err)
	}
	onTeam := false
	for _, row := range placements {
		if row.UserID == userID && row.TeamNumber != nil {
			onTeam = true
			break
		}
	}
	if !onTeam {
		return &AbsenceValidationError{Message: "Only players on a team can be marked absent"}
	}

	if err := s.q.UpsertNoShow(ctx, db.UpsertNoShowParams{
		EventID: eventID,
		UserID:  userID,
		HostID:  meta.OwnerID,
	}); err != nil {
		return fmt.Errorf("upsert no-show: %w", err)
	}
	return nil
}

// visibleAbsentUserIDs filters a game's no-show list to what the viewer may see:
// the host sees every mark, anyone else sees only their own.
func visibleAbsentUserIDs(absentIDs []uuid.UUID, viewerID uuid.UUID, viewerIsHost bool) []uuid.UUID {
	if viewerIsHost {
		if absentIDs == nil {
			return []uuid.UUID{}
		}
		return absentIDs
	}
	for _, id := range absentIDs {
		if id == viewerID {
			return []uuid.UUID{viewerID}
		}
	}
	return []uuid.UUID{}
}
