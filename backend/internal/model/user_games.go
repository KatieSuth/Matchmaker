package model

import (
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

type UserGame struct {
	UserID          uuid.UUID  `json:"user_id"`
	GameID          uuid.UUID  `json:"game_id"`
	GameName        string     `json:"game_name"`
	InGameName      *string    `json:"in_game_name"`
	CurrentRank     *uuid.UUID `json:"current_rank"`
	CurrentRankName string     `json:"current_rank_name"`
	PeakRank        *uuid.UUID `json:"peak_rank"`
	PeakRankName    string     `json:"peak_rank_name"`
	ShowRank        bool       `json:"show_rank"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

/*
 * Mapping functions to enable json mappings (sqlc can do
 * this but handling it manually gives more flexibility)
 */
func MapDbGetGamesForUserRowToUserGame(row db.GetGamesForUserRow) UserGame {
	return UserGame{
		UserID:          row.UserID,
		GameID:          row.GameID,
		GameName:        row.GameName,
		InGameName:      &row.InGameName,
		CurrentRank:     row.CurrentRank,
		CurrentRankName: row.CurrentRankName,
		PeakRank:        row.PeakRank,
		PeakRankName:    row.PeakRankName,
		ShowRank:        row.ShowRank,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func MapDbUserGamesToUserGames(dbUserGames []db.GetGamesForUserRow) []UserGame {
	userGames := make([]UserGame, 0, len(dbUserGames))
	for _, row := range dbUserGames {
		userGames = append(userGames, MapDbGetGamesForUserRowToUserGame(row))
	}
	return userGames
}

func MapDbUserGameToUserGame(dbUserGame db.UserGame) UserGame {
	userGame := UserGame{
		UserID:      dbUserGame.UserID,
		GameID:      dbUserGame.GameID,
		InGameName:  &dbUserGame.InGameName,
		CurrentRank: dbUserGame.CurrentRank,
		PeakRank:    dbUserGame.PeakRank,
		ShowRank:    dbUserGame.ShowRank,
		CreatedAt:   dbUserGame.CreatedAt,
		UpdatedAt:   dbUserGame.UpdatedAt,
	}

	return userGame
}
