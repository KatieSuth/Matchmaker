package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrInvalidPlayerSwap is the sentinel for swap validation failures surfaced to handlers.
var ErrInvalidPlayerSwap = errors.New("invalid player swap")

// SwapValidationError carries a client-facing message for swap validation failures.
type SwapValidationError struct {
	Message string
}

func (e *SwapValidationError) Error() string {
	return e.Message
}

func (e *SwapValidationError) Unwrap() error {
	return ErrInvalidPlayerSwap
}

type swapPlacement struct {
	placed     bool
	lobbyID    uuid.UUID
	teamNumber *int32
}

type swapRegistration struct {
	canSubstitute bool
}

// SwapPlayersForEvent exchanges two players' placements for one locked-in game.
// Host authorization, validation, and transactional persistence run here; affected lobbies
// get refreshed hosts and live fairness_warning values (fairness_warning_at_lock is unchanged).
func (s *PostgresStore) SwapPlayersForEvent(
	ctx context.Context,
	eventID, ownerID, userA, userB uuid.UUID,
	settings matchmaking.Settings,
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

	regA, err := s.loadSwapRegistration(ctx, eventID, userA)
	if err != nil {
		return err
	}
	regB, err := s.loadSwapRegistration(ctx, eventID, userB)
	if err != nil {
		return err
	}

	placements, err := s.q.GetPlayerPlacementsForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("get player placements: %w", err)
	}
	placementByUser := make(map[uuid.UUID]swapPlacement, len(placements))
	for _, row := range placements {
		placementByUser[row.UserID] = swapPlacement{
			placed:     true,
			lobbyID:    row.LobbyID,
			teamNumber: row.TeamNumber,
		}
	}

	placeA, okA := placementByUser[userA]
	if !okA {
		placeA = swapPlacement{placed: false}
	}
	placeB, okB := placementByUser[userB]
	if !okB {
		placeB = swapPlacement{placed: false}
	}

	if userA == userB {
		return &SwapValidationError{Message: "Cannot swap a player with themselves"}
	}
	if !placeA.placed && !placeB.placed {
		return &SwapValidationError{Message: "Both players are unplaced"}
	}
	if sameTeamPlacement(placeA, placeB) {
		return &SwapValidationError{Message: "Cannot swap players on the same team"}
	}
	if sameSubPoolPlacement(placeA, placeB) {
		return &SwapValidationError{Message: "Cannot swap players in the same sub pool"}
	}
	if isSubUnplacedSwap(placeA, placeB) {
		return &SwapValidationError{Message: "Cannot swap between substitutes and unplaced players"}
	}

	newA := resolveSwapDestination(placeA, placeB, regA.canSubstitute)
	newB := resolveSwapDestination(placeB, placeA, regB.canSubstitute)

	lobbyRows, err := s.q.GetLobbiesForEvent(ctx, &eventID)
	if err != nil {
		return fmt.Errorf("get lobbies for event: %w", err)
	}
	if err := validateSubMinimumAfterSwap(placementByUser, userA, userB, newA, newB, lobbyRows, int(meta.SubMin)); err != nil {
		return err
	}

	affectedLobbyIDs := affectedLobbySet(placeA, placeB, newA, newB)

	return s.WithTx(ctx, func(tx Store) error {
		txStore, ok := tx.(*PostgresStore)
		if !ok {
			return fmt.Errorf("unexpected tx store type %T", tx)
		}

		if placeA.placed {
			if err := txStore.q.DeletePlayer(ctx, db.DeletePlayerParams{
				LobbyID: placeA.lobbyID,
				UserID:  userA,
			}); err != nil {
				return fmt.Errorf("delete player A: %w", err)
			}
		}
		if placeB.placed {
			if err := txStore.q.DeletePlayer(ctx, db.DeletePlayerParams{
				LobbyID: placeB.lobbyID,
				UserID:  userB,
			}); err != nil {
				return fmt.Errorf("delete player B: %w", err)
			}
		}

		if newA.placed {
			if err := txStore.q.CreatePlayer(ctx, db.CreatePlayerParams{
				LobbyID:    newA.lobbyID,
				UserID:     userA,
				TeamNumber: newA.teamNumber,
			}); err != nil {
				return fmt.Errorf("insert player A: %w", err)
			}
		}
		if newB.placed {
			if err := txStore.q.CreatePlayer(ctx, db.CreatePlayerParams{
				LobbyID:    newB.lobbyID,
				UserID:     userB,
				TeamNumber: newB.teamNumber,
			}); err != nil {
				return fmt.Errorf("insert player B: %w", err)
			}
		}

		for lobbyID := range affectedLobbyIDs {
			if err := txStore.recomputeLobbyAfterSwap(ctx, lobbyID, settings, meta.GameModeID); err != nil {
				return err
			}
		}
		return nil
	})
}

// loadSwapRegistration loads can_substitute for a registered player, required to resolve sub-slot destinations.
func (s *PostgresStore) loadSwapRegistration(ctx context.Context, eventID, userID uuid.UUID) (swapRegistration, error) {
	row, err := s.q.GetRegistrationByEventAndUser(ctx, db.GetRegistrationByEventAndUserParams{
		EventID: eventID,
		UserID:  userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return swapRegistration{}, &SwapValidationError{Message: "One or both players are not registered for this game"}
		}
		return swapRegistration{}, fmt.Errorf("get registration: %w", err)
	}
	return swapRegistration{canSubstitute: row.CanSubstitute}, nil
}

// sameTeamPlacement reports whether both players occupy the same roster slot in one lobby.
func sameTeamPlacement(a, b swapPlacement) bool {
	if !a.placed || !b.placed {
		return false
	}
	if a.lobbyID != b.lobbyID {
		return false
	}
	if a.teamNumber == nil || b.teamNumber == nil {
		return false
	}
	return *a.teamNumber == *b.teamNumber
}

// sameSubPoolPlacement reports whether both players are substitutes for this game.
// Substitutes are treated as one per-game pool in the product UI, even though rows are stored per lobby.
func sameSubPoolPlacement(a, b swapPlacement) bool {
	if !a.placed || !b.placed {
		return false
	}
	return a.teamNumber == nil && b.teamNumber == nil
}

// resolveSwapDestination maps the counterpart's slot onto this player.
// Non-substitute players landing in a sub slot become unplaced instead.
// Roster players displaced by an unplaced player move into their lobby's sub
// pool when they can substitute; otherwise they become unplaced.
func resolveSwapDestination(ownPrevious, counterpart swapPlacement, canSubstitute bool) swapPlacement {
	if !counterpart.placed {
		if ownPrevious.placed && ownPrevious.teamNumber != nil && canSubstitute {
			return swapPlacement{placed: true, lobbyID: ownPrevious.lobbyID, teamNumber: nil}
		}
		return swapPlacement{placed: false}
	}
	if counterpart.teamNumber == nil && !canSubstitute {
		return swapPlacement{placed: false}
	}
	return swapPlacement{
		placed:     true,
		lobbyID:    counterpart.lobbyID,
		teamNumber: counterpart.teamNumber,
	}
}

// validateSubMinimumAfterSwap rejects swaps that would drop a lobby below sub_min.
// Skipped for single-lobby games or when sub_min is zero.
func validateSubMinimumAfterSwap(
	before map[uuid.UUID]swapPlacement,
	userA, userB uuid.UUID,
	newA, newB swapPlacement,
	lobbies []db.GetLobbiesForEventRow,
	subMin int,
) error {
	if len(lobbies) < 2 || subMin <= 0 {
		return nil
	}

	after := make(map[uuid.UUID]swapPlacement, len(before))
	for userID, placement := range before {
		after[userID] = placement
	}
	delete(after, userA)
	delete(after, userB)
	if newA.placed {
		after[userA] = newA
	}
	if newB.placed {
		after[userB] = newB
	}

	subsPerLobby := make(map[uuid.UUID]int)
	for _, placement := range after {
		if !placement.placed || placement.teamNumber != nil {
			continue
		}
		subsPerLobby[placement.lobbyID]++
	}

	for _, lobby := range lobbies {
		if subsPerLobby[lobby.ID] < subMin {
			return &TeamCreationError{
				Sentinel: ErrInsufficientSubstitutes,
				Message:  "This swap would leave fewer than the required substitutes in a lobby",
			}
		}
	}
	return nil
}

// affectedLobbySet returns lobby IDs whose host and fairness state must be recomputed after a swap.
func affectedLobbySet(placeA, placeB, newA, newB swapPlacement) map[uuid.UUID]bool {
	ids := make(map[uuid.UUID]bool)
	for _, p := range []swapPlacement{placeA, placeB, newA, newB} {
		if p.placed {
			ids[p.lobbyID] = true
		}
	}
	return ids
}

// recomputeLobbyAfterSwap refreshes lobby host and live fairness_warning after roster changes.
// fairness_warning_at_lock is left unchanged so the UI can distinguish lock-in vs manual-edit warnings.
func (s *PostgresStore) recomputeLobbyAfterSwap(
	ctx context.Context,
	lobbyID uuid.UUID,
	settings matchmaking.Settings,
	gameModeID uuid.UUID,
) error {
	playerRows, err := s.q.GetPlayersForLobby(ctx, db.GetPlayersForLobbyParams{
		ViewerIsHost: true,
		LobbyID:      lobbyID,
	})
	if err != nil {
		return fmt.Errorf("get players for lobby %s: %w", lobbyID.String(), err)
	}

	lobbyPlan := matchmaking.LobbyPlan{
		Roster: make([]matchmaking.Player, 0),
		Subs:   make([]matchmaking.Player, 0),
	}
	for _, row := range playerRows {
		p := matchmaking.Player{
			UserID:        row.UserID,
			AvgRank:       float64(row.AvgRankOrder),
			CanSubstitute: row.CanSubstitute,
			CanLobbyHost:  row.CanLobbyHost,
			CreatedAt:     row.CreatedAt,
		}
		if row.TeamNumber == nil {
			lobbyPlan.Subs = append(lobbyPlan.Subs, p)
			continue
		}
		teamNum := int(*row.TeamNumber)
		p.TeamNumber = &teamNum
		lobbyPlan.Roster = append(lobbyPlan.Roster, p)
	}

	hostID := matchmaking.PickLobbyHost(lobbyPlan)
	if err := s.q.UpdateLobbyHost(ctx, db.UpdateLobbyHostParams{
		ID:   lobbyID,
		Host: hostID,
	}); err != nil {
		return fmt.Errorf("update lobby host: %w", err)
	}

	mode, err := s.q.GetGameModeById(ctx, gameModeID)
	if err != nil {
		return fmt.Errorf("get game mode: %w", err)
	}
	tierCount, err := s.q.GetMaxRankOrderForGame(ctx, &mode.GameID)
	if err != nil {
		return fmt.Errorf("get max rank order: %w", err)
	}

	fairnessWarning := matchmaking.IsLobbyUnfair(lobbyPlan, settings, int(tierCount))
	if err := s.q.UpdateLobbyFairnessWarning(ctx, db.UpdateLobbyFairnessWarningParams{
		ID:              lobbyID,
		FairnessWarning: fairnessWarning,
	}); err != nil {
		return fmt.Errorf("update lobby fairness: %w", err)
	}
	return nil
}
