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
func MapDbUserGamesToUserGames(dbUserGames []db.GetGamesForUserRow) []UserGame {
	userGames := []UserGame{}

	for _, dbUserGame := range dbUserGames {
		userGames = append(userGames, UserGame{
			UserID:          dbUserGame.UserID,
			GameID:          dbUserGame.GameID,
			GameName:        dbUserGame.GameName,
			InGameName:      &dbUserGame.InGameName,
			CurrentRank:     dbUserGame.CurrentRank,
			CurrentRankName: dbUserGame.CurrentRankName,
			PeakRank:        dbUserGame.PeakRank,
			PeakRankName:    dbUserGame.PeakRankName,
			ShowRank:        dbUserGame.ShowRank,
			CreatedAt:       dbUserGame.CreatedAt,
			UpdatedAt:       dbUserGame.UpdatedAt,
		})
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
