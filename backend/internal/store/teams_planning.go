package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/google/uuid"
)

var (
	ErrInsufficientPlayers     = errors.New("insufficient players")
	ErrInsufficientSubstitutes = errors.New("insufficient substitutes")
)

// TeamCreationError wraps a sentinel with a client-facing message for handler mapping.
type TeamCreationError struct {
	Sentinel error
	Message  string
}

func (e *TeamCreationError) Error() string {
	return e.Message
}

func (e *TeamCreationError) Unwrap() error {
	return e.Sentinel
}

// buildMatchmakingConfig assembles per-game matchmaking input from group settings and game mode.
func buildMatchmakingConfig(event db.Event, group db.EventGroup, mode modelGameModeInfo, tierCount int, gameIndex int) matchmaking.Config {
	slots := int(mode.TeamSize) * 2
	label := fmt.Sprintf("Game %d (%s)", gameIndex+1, mode.Name)
	return matchmaking.Config{
		EventID:   event.ID,
		TeamSize:  int(mode.TeamSize),
		SubMin:    int(group.SubMin),
		SortLogic: group.SortLogic,
		TierCount: tierCount,
		GameLabel: label,
		Slots:     slots,
	}
}

type modelGameModeInfo struct {
	Name     string
	TeamSize int32
	GameID   uuid.UUID
}

// buildGroupRegistrationCounts returns how many games in the group each user signed up for (tie-break input).
func (s *PostgresStore) buildGroupRegistrationCounts(ctx context.Context, groupID uuid.UUID) (map[uuid.UUID]int, error) {
	events, err := s.q.GetEventsByGroupId(ctx, &groupID)
	if err != nil {
		return nil, fmt.Errorf("get events by group: %w", err)
	}

	counts := make(map[uuid.UUID]int)
	for _, event := range events {
		regs, err := s.q.GetRegistrationsForEvent(ctx, event.ID)
		if err != nil {
			return nil, fmt.Errorf("get registrations for event %s: %w", event.ID.String(), err)
		}
		for _, reg := range regs {
			counts[reg.UserID]++
		}
	}
	return counts, nil
}

// mapRegistrationsToPlayers converts DB registration rows into matchmaking.Player values.
func mapRegistrationsToPlayers(rows []db.GetMatchmakingRegistrationsForEventRow, regCounts map[uuid.UUID]int) []matchmaking.Player {
	players := make([]matchmaking.Player, 0, len(rows))
	for _, row := range rows {
		players = append(players, matchmaking.Player{
			UserID:              row.UserID,
			AvgRank:             matchmaking.AverageRankOrder(int(row.CurrentRankOrder), int(row.PeakRankOrder)),
			CanSubstitute:       row.CanSubstitute,
			CanLobbyHost:        row.CanLobbyHost,
			RegisteredGameCount: regCounts[row.UserID],
			CreatedAt:           row.CreatedAt,
		})
	}
	return players
}

// planTeamsForGroup runs matchmaking independently for each game in the group before persistence.
func (s *PostgresStore) planTeamsForGroup(ctx context.Context, group db.EventGroup, settings matchmaking.Settings) ([]matchmaking.GamePlan, error) {
	events, err := s.q.GetEventsByGroupId(ctx, &group.ID)
	if err != nil {
		return nil, fmt.Errorf("get events by group: %w", err)
	}

	regCounts, err := s.buildGroupRegistrationCounts(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	plans := make([]matchmaking.GamePlan, 0, len(events))
	for i, eventRow := range events {
		regRows, err := s.q.GetMatchmakingRegistrationsForEvent(ctx, eventRow.ID)
		if err != nil {
			return nil, fmt.Errorf("get matchmaking registrations for event %s: %w", eventRow.ID.String(), err)
		}
		if len(regRows) == 0 {
			continue
		}

		mode, err := s.q.GetGameModeById(ctx, eventRow.GameModeID)
		if err != nil {
			return nil, fmt.Errorf("get game mode for event %s: %w", eventRow.ID.String(), err)
		}

		tierCount, err := s.q.GetMaxRankOrderForGame(ctx, &mode.GameID)
		if err != nil {
			return nil, fmt.Errorf("get max rank order for game: %w", err)
		}

		cfg := buildMatchmakingConfig(eventRow, group, modelGameModeInfo{
			Name:     mode.Name,
			TeamSize: mode.TeamSize,
			GameID:   mode.GameID,
		}, int(tierCount), i)

		players := mapRegistrationsToPlayers(regRows, regCounts)
		plan, err := matchmaking.PlanEvent(players, cfg, settings)
		if err != nil {
			var valErr *matchmaking.ValidationError
			if errors.As(err, &valErr) {
				return nil, &TeamCreationError{
					Sentinel: ErrInsufficientPlayers,
					Message:  valErr.Message,
				}
			}
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

func (s *PostgresStore) persistTeamPlans(ctx context.Context, plans []matchmaking.GamePlan) error {
	for _, plan := range plans {
		for _, lobby := range plan.Lobbies {
			created, err := s.q.CreateLobby(ctx, db.CreateLobbyParams{
				EventID:         &plan.EventID,
				Host:            lobby.HostID,
				FairnessWarning: lobby.FairnessWarning,
			})
			if err != nil {
				return fmt.Errorf("create lobby for event %s: %w", plan.EventID.String(), err)
			}

			for _, p := range lobby.Roster {
				err = s.q.CreatePlayer(ctx, db.CreatePlayerParams{
					LobbyID:    created.ID,
					UserID:     p.UserID,
					TeamNumber: intPtrToInt32Ptr(p.TeamNumber),
				})
				if err != nil {
					return fmt.Errorf("create roster player for event %s: %w", plan.EventID.String(), err)
				}
			}
			for _, p := range lobby.Subs {
				err = s.q.CreatePlayer(ctx, db.CreatePlayerParams{
					LobbyID:    created.ID,
					UserID:     p.UserID,
					TeamNumber: nil,
				})
				if err != nil {
					return fmt.Errorf("create sub player for event %s: %w", plan.EventID.String(), err)
				}
			}
		}
	}
	return nil
}

func intPtrToInt32Ptr(v *int) *int32 {
	if v == nil {
		return nil
	}
	n := int32(*v)
	return &n
}
