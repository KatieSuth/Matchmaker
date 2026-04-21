package model

import (
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

type GameMode struct {
	ID        uuid.UUID  `json:"id"`
	GameID    uuid.UUID  `json:"game_id"`
	Name      string     `json:"name"`
	TeamSize  int32      `json:"team_size"`
	OwnerID   *uuid.UUID `json:"owner_id"`
	Duration  int32      `json:"duration"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func MapDbGameModeToGameMode(dbGameMode db.GameMode) GameMode {
	return GameMode{
		ID:        dbGameMode.ID,
		GameID:    dbGameMode.GameID,
		Name:      dbGameMode.Name,
		TeamSize:  dbGameMode.TeamSize,
		OwnerID:   dbGameMode.OwnerID,
		Duration:  dbGameMode.Duration,
		CreatedAt: dbGameMode.CreatedAt,
		UpdatedAt: dbGameMode.UpdatedAt,
	}
}

func MapDbGameModesToGameModes(dbGameModes []db.GameMode) []GameMode {
	gameModes := make([]GameMode, 0, len(dbGameModes))
	for _, dbGameMode := range dbGameModes {
		gameModes = append(gameModes, MapDbGameModeToGameMode(dbGameMode))
	}
	return gameModes
}
