package model

import (
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

type Game struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	OwnerID   *uuid.UUID `json:"owner_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

/*
 * Mapping function to enable json mappings (sqlc can do
 * this but handling it manually gives more flexibility)
 */
func MapDbGameToGame(dbGames []db.Game) []Game {
	games := []Game{}

	for _, dbGame := range dbGames {
		games = append(games, Game{
			ID:        dbGame.ID,
			Name:      dbGame.Name,
			OwnerID:   dbGame.OwnerID,
			CreatedAt: dbGame.CreatedAt,
			UpdatedAt: dbGame.UpdatedAt,
		})
	}

	return games
}
