package model

import (
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

type GameRank struct {
	ID        uuid.UUID  `json:"id"`
	GameID    *uuid.UUID `json:"game_id"`
	Name      string     `json:"name"`
	Order     int32      `json:"order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

/*
 * Mapping functions to enable json mappings (sqlc can do
 * this but handling it manually gives more flexibility)
 */
func MapDbGameRankToGameRank(dbGameRank db.GameRank) GameRank {
	return GameRank{
		ID:        dbGameRank.ID,
		GameID:    dbGameRank.GameID,
		Name:      dbGameRank.Name,
		Order:     dbGameRank.Order,
		CreatedAt: dbGameRank.CreatedAt,
		UpdatedAt: dbGameRank.UpdatedAt,
	}
}

func MapDbGameRanksToGameRanks(dbGameRanks []db.GameRank) []GameRank {
	gameRanks := make([]GameRank, 0, len(dbGameRanks))
	for _, dbGameRank := range dbGameRanks {
		gameRanks = append(gameRanks, MapDbGameRankToGameRank(dbGameRank))
	}
	return gameRanks
}
