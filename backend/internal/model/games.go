package model

import (
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

type Game struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	OwnerID      *uuid.UUID `json:"owner_id"`
	JoinLinkBase *string    `json:"join_link_base"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

/*
 * Mapping functions to enable json mappings (sqlc can do
 * this but handling it manually gives more flexibility)
 */
func MapDbGameToGame(dbGame db.Game) Game {
	return Game{
		ID:           dbGame.ID,
		Name:         dbGame.Name,
		OwnerID:      dbGame.OwnerID,
		JoinLinkBase: dbGame.JoinLinkBase,
		CreatedAt:    dbGame.CreatedAt,
		UpdatedAt:    dbGame.UpdatedAt,
	}
}

func MapDbGamesToGames(dbGames []db.Game) []Game {
	games := make([]Game, 0, len(dbGames))
	for _, dbGame := range dbGames {
		games = append(games, MapDbGameToGame(dbGame))
	}
	return games
}
