package store

import (
	"context"
	"fmt"
	"sort"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
)

// sortLobbyPlayersByRankDesc orders team players highest average rank first for API responses.
func sortLobbyPlayersByRankDesc(players []model.LobbyPlayer) []model.LobbyPlayer {
	sorted := append([]model.LobbyPlayer(nil), players...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].AvgRankOrder != sorted[j].AvgRankOrder {
			return sorted[i].AvgRankOrder > sorted[j].AvgRankOrder
		}
		return sorted[i].DiscordName < sorted[j].DiscordName
	})
	return sorted
}

// loadEventLobbiesAndUnplaced builds lobby team/sub assignments and derives unplaced registrations.
func (s *PostgresStore) loadEventLobbiesAndUnplaced(
	ctx context.Context,
	eventID uuid.UUID,
	viewerIsHost bool,
	registrations []model.EventRegistration,
) ([]model.EventLobby, []model.EventRegistration, error) {
	lobbyRows, err := s.q.GetLobbiesForEvent(ctx, &eventID)
	if err != nil {
		return nil, nil, fmt.Errorf("get lobbies for event: %w", err)
	}

	placedUsers := make(map[uuid.UUID]bool)
	lobbies := make([]model.EventLobby, 0, len(lobbyRows))

	for _, lobbyRow := range lobbyRows {
		playerRows, err := s.q.GetPlayersForLobby(ctx, db.GetPlayersForLobbyParams{
			ViewerIsHost: viewerIsHost,
			LobbyID:      lobbyRow.ID,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("get players for lobby %s: %w", lobbyRow.ID.String(), err)
		}

		lobby := model.MapDbGetLobbiesForEventRowToEventLobby(lobbyRow)
		lobby.Teams = []model.EventTeam{
			{TeamNumber: 1, Players: []model.LobbyPlayer{}},
			{TeamNumber: 2, Players: []model.LobbyPlayer{}},
		}

		for _, pr := range playerRows {
			placedUsers[pr.UserID] = true
			player := model.MapDbGetPlayersForLobbyRowToLobbyPlayer(pr)
			if pr.TeamNumber == nil {
				lobby.Subs = append(lobby.Subs, player)
				continue
			}
			switch *pr.TeamNumber {
			case 1:
				lobby.Teams[0].Players = append(lobby.Teams[0].Players, player)
			case 2:
				lobby.Teams[1].Players = append(lobby.Teams[1].Players, player)
			}
		}
		for i := range lobby.Teams {
			lobby.Teams[i].Players = sortLobbyPlayersByRankDesc(lobby.Teams[i].Players)
		}
		lobbies = append(lobbies, lobby)
	}

	var unplaced []model.EventRegistration
	if viewerIsHost {
		for _, reg := range registrations {
			if !placedUsers[reg.UserID] {
				unplaced = append(unplaced, reg)
			}
		}
	}
	if unplaced == nil {
		unplaced = []model.EventRegistration{}
	}

	return lobbies, unplaced, nil
}
