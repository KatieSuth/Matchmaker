// Package model defines JSON-facing domain types and maps sqlc-generated row types
// (internal/db) into those shapes, keeping the DB layer from leaking into handlers.
package model

import (
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

// DashboardEvent is a single card on the "my events" list (keyset-paginated).
type DashboardEvent struct {
	ID               uuid.UUID `json:"id"`
	GameName         string    `json:"game_name"`
	GameMode         string    `json:"game_mode"`
	EventDate        time.Time `json:"event_date"`
	HostID           uuid.UUID `json:"host_id"`
	HostName         string    `json:"host_name"`
	RegisteredCount  int       `json:"registered_count"`
	RegistrationOpen bool      `json:"registration_open"`
}

// EventRegistration is one sign-up for an event, as shown on the event detail page.
type EventRegistration struct {
	EventID         uuid.UUID `json:"event_id"`
	UserID          uuid.UUID `json:"user_id"`
	DiscordName     string    `json:"discord_name"`
	InGameName      string    `json:"in_game_name"`
	Pronouns        string    `json:"pronouns"`
	CurrentRankName string    `json:"current_rank_name"`
	PeakRankName    string    `json:"peak_rank_name"`
	AvgRankName     string    `json:"avg_rank_name"`
	CanSubstitute   bool      `json:"can_substitute"`
	CanLobbyHost    bool      `json:"can_lobby_host"`
	DuoRequest      *string   `json:"duo_request"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// LobbyPlayer is a player shown on a formed team or sub list.
type LobbyPlayer struct {
	UserID           uuid.UUID `json:"user_id"`
	DiscordName      string    `json:"discord_name"`
	InGameName       string    `json:"in_game_name"`
	Pronouns         string    `json:"pronouns"`
	CurrentRankName  string `json:"current_rank_name"`
	CurrentRankOrder int    `json:"current_rank_order"`
	PeakRankName     string `json:"peak_rank_name"`
	PeakRankOrder    int    `json:"peak_rank_order"`
	AvgRankName      string `json:"avg_rank_name"`
	AvgRankOrder     int    `json:"avg_rank_order"`
	CanSubstitute    bool      `json:"can_substitute"`
	CanLobbyHost     bool      `json:"can_lobby_host"`
	DuoRequest       *string   `json:"duo_request"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// EventTeam is one side within a lobby.
type EventTeam struct {
	TeamNumber int           `json:"team_number"`
	Players    []LobbyPlayer `json:"players"`
}

// EventLobby is one match lobby for a game after lock-in.
type EventLobby struct {
	ID                    uuid.UUID   `json:"id"`
	HostID                *uuid.UUID  `json:"host_id"`
	FairnessWarning       bool        `json:"fairness_warning"`
	FairnessWarningAtLock bool        `json:"fairness_warning_at_lock"`
	Teams                 []EventTeam `json:"teams"`
	Subs                  []LobbyPlayer `json:"subs"`
}

// EventGroupEvent is one scheduled game within a group, with optional registration rows.
type EventGroupEvent struct {
	ID               uuid.UUID           `json:"id"`
	StartTime        time.Time           `json:"start_time"`
	GameModeID       uuid.UUID           `json:"game_mode_id"`
	GameModeName     string              `json:"game_mode_name"`
	TeamSize         int                 `json:"team_size"`
	RegisteredCount  int                 `json:"registered_count"`
	LobbiesCount     int                 `json:"lobbies_count"`
	PlayerRegistered bool                `json:"player_registered"`
	Registrations    []EventRegistration `json:"registrations"`
	Lobbies          []EventLobby        `json:"lobbies"`
	Unplaced         []EventRegistration `json:"unplaced"`
}

// EventGroupDetail is the full host/participant view for a group: header plus each game.
type EventGroupDetail struct {
	ID               uuid.UUID         `json:"id"`
	OwnerID          uuid.UUID         `json:"owner_id"`
	OwnerName        string            `json:"owner_name"`
	OwnerPronouns    string            `json:"owner_pronouns"`
	GameModeID       uuid.UUID         `json:"game_mode_id"`
	GameModeName     string            `json:"game_mode_name"`
	GameID           uuid.UUID         `json:"game_id"`
	GameName         string            `json:"game_name"`
	TeamSize         int               `json:"team_size"`
	SubMin           int               `json:"sub_min"`
	RegistrationOpen bool              `json:"registration_open"`
	Region           string            `json:"region"`
	SortLogic        string            `json:"sort_logic"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	Events           []EventGroupEvent `json:"events"`
}

/*
 * Mapping functions to enable json mappings (sqlc can do
 * this but handling it manually gives more flexibility)
 */
func MapDbGetEventsForUserRowToDashboardEvent(row db.GetEventsForUserRow) DashboardEvent {
	return DashboardEvent{
		ID:               row.ID,
		GameName:         row.GameName,
		GameMode:         row.GameMode,
		EventDate:        row.EventDate,
		HostID:           row.HostID,
		HostName:         row.HostName,
		RegisteredCount:  int(row.RegisteredCount),
		RegistrationOpen: row.RegistrationOpen,
	}
}

func MapDbGetRegistrationDataByEventIdRowToEventRegistration(row db.GetRegistrationDataByEventIdRow) EventRegistration {
	discordName := ""
	if row.DiscordName != nil {
		discordName = *row.DiscordName
	}
	return EventRegistration{
		EventID:         row.EventID,
		UserID:          row.UserID,
		DiscordName:     discordName,
		InGameName:      row.InGameName,
		Pronouns:        row.Pronouns,
		CurrentRankName: row.CurrentRankName,
		PeakRankName:    row.PeakRankName,
		AvgRankName:     row.AvgRankName,
		CanSubstitute:   row.CanSubstitute,
		CanLobbyHost:    row.CanLobbyHost,
		DuoRequest:      row.DuoRequest,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func MapDbGetRegistrationDataByEventIdRowsToEventRegistrations(rows []db.GetRegistrationDataByEventIdRow) []EventRegistration {
	out := make([]EventRegistration, 0, len(rows))
	for _, row := range rows {
		out = append(out, MapDbGetRegistrationDataByEventIdRowToEventRegistration(row))
	}
	return out
}

func MapDbGetGroupEventsSummaryRowToEventGroupEvent(row db.GetGroupEventsSummaryRow, registrations []EventRegistration) EventGroupEvent {
	return EventGroupEvent{
		ID:               row.ID,
		StartTime:        row.StartTime,
		GameModeID:       row.GameModeID,
		GameModeName:     row.GameModeName,
		TeamSize:         int(row.TeamSize),
		RegisteredCount:  int(row.RegisteredCount),
		LobbiesCount:     int(row.LobbiesCount),
		PlayerRegistered: row.PlayerRegistered,
		Registrations:    registrations,
		Lobbies:          []EventLobby{},
		Unplaced:         []EventRegistration{},
	}
}

func MapDbGetLobbiesForEventRowToEventLobby(row db.GetLobbiesForEventRow) EventLobby {
	return EventLobby{
		ID:                    row.ID,
		HostID:                row.Host,
		FairnessWarning:       row.FairnessWarning,
		FairnessWarningAtLock: row.FairnessWarningAtLock,
		Teams:                 []EventTeam{},
		Subs:                  []LobbyPlayer{},
	}
}

func MapDbGetPlayersForLobbyRowToLobbyPlayer(row db.GetPlayersForLobbyRow) LobbyPlayer {
	discordName := ""
	if row.DiscordName != nil {
		discordName = *row.DiscordName
	}
	return LobbyPlayer{
		UserID:           row.UserID,
		DiscordName:      discordName,
		InGameName:       row.InGameName,
		Pronouns:         row.Pronouns,
		CurrentRankName:  row.CurrentRankName,
		CurrentRankOrder: int(row.CurrentRankOrder),
		PeakRankName:     row.PeakRankName,
		PeakRankOrder:    int(row.PeakRankOrder),
		AvgRankName:      row.AvgRankName,
		AvgRankOrder:     int(row.AvgRankOrder),
		CanSubstitute:    row.CanSubstitute,
		CanLobbyHost:     row.CanLobbyHost,
		DuoRequest:       row.DuoRequest,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func MapDbGetEventGroupDetailByIdRowToEventGroupDetail(row db.GetEventGroupDetailByIdRow, events []EventGroupEvent) EventGroupDetail {
	return EventGroupDetail{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
		OwnerName:        row.OwnerName,
		OwnerPronouns:    row.OwnerPronouns,
		GameModeID:       row.GameModeID,
		GameModeName:     row.GameModeName,
		GameID:           row.GameID,
		GameName:         row.GameName,
		TeamSize:         int(row.TeamSize),
		SubMin:           int(row.SubMin),
		RegistrationOpen: row.RegistrationOpen,
		Region:           row.Region,
		SortLogic:        row.SortLogic,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
		Events:           events,
	}
}
