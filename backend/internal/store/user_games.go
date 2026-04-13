package store

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) GetUserGamesForUser(ctx context.Context, userID uuid.UUID) ([]model.UserGame, error) {
	dbUserGames, err := s.q.GetGamesForUser(ctx, userID)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return []model.UserGame{}, fmt.Errorf("looking up user's games: %w", err)
	}

	return model.MapDbUserGamesToUserGames(dbUserGames), nil
}

func (s *PostgresStore) UpsertGameForUser(ctx context.Context, userID uuid.UUID, ug model.UserGame) (model.UserGame, int, error) {
	/*** validate provided info ***/

	//make sure game exists
	_, err := s.q.GetGameById(ctx, ug.GameID)
	if err != nil {
		return model.UserGame{}, http.StatusBadRequest, fmt.Errorf("Invalid game: %s", ug.GameID.String())
	}

	//make sure we got existing ranks
	if ug.CurrentRank == nil || *ug.CurrentRank == uuid.Nil {
		return model.UserGame{}, http.StatusBadRequest, fmt.Errorf("Current rank must not be empty")
	}

	if ug.PeakRank == nil || *ug.PeakRank == uuid.Nil {
		return model.UserGame{}, http.StatusBadRequest, fmt.Errorf("Peak rank must not be empty")
	}

	currentRank, err := s.q.GetRankById(ctx, *ug.CurrentRank)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return model.UserGame{}, http.StatusInternalServerError, fmt.Errorf("looking up current rank by ID: %w", err)
	} else if err != nil {
		return model.UserGame{}, http.StatusBadRequest, fmt.Errorf("Invalid current rank: %s", ug.CurrentRank.String())
	}

	peakRank, err := s.q.GetRankById(ctx, *ug.PeakRank)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return model.UserGame{}, http.StatusInternalServerError, fmt.Errorf("looking up peak rank by ID: %w", err)
	} else if err != nil {
		return model.UserGame{}, http.StatusBadRequest, fmt.Errorf("invalid peak rank: %s", ug.PeakRank.String())
	}

	//make sure ranks are logical
	if currentRank.Order > peakRank.Order {
		return model.UserGame{}, http.StatusBadRequest, fmt.Errorf("Peak rank must be greater than or equal to current rank")
	}

	/*** Upsert ***/
	_, err = s.q.GetGameForUserByIds(ctx, db.GetGameForUserByIdsParams{
		UserID: userID,
		GameID: ug.GameID,
	})

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		// Genuine DB error — surface it
		return model.UserGame{}, http.StatusInternalServerError, fmt.Errorf("looking up user game: %w", err)
	}

	var dbUserGame db.UserGame
	if errors.Is(err, pgx.ErrNoRows) {
		//userGame link doesn't exist, create it
		dbUserGame, err = s.q.CreateGameForUser(ctx, db.CreateGameForUserParams{
			UserID:      userID,
			GameID:      ug.GameID,
			InGameName:  *ug.InGameName,
			CurrentRank: ug.CurrentRank,
			PeakRank:    ug.PeakRank,
			ShowRank:    ug.ShowRank,
		})
		if err != nil {
			return model.UserGame{}, http.StatusInternalServerError, fmt.Errorf("creating game for user: %w", err)
		}
	} else {
		//userGame link exists, update it
		dbUserGame, err = s.q.UpdateGameForUser(ctx, db.UpdateGameForUserParams{
			UserID:      userID,
			GameID:      ug.GameID,
			InGameName:  *ug.InGameName,
			CurrentRank: ug.CurrentRank,
			PeakRank:    ug.PeakRank,
			ShowRank:    ug.ShowRank,
		})

		if err != nil {
			return model.UserGame{}, http.StatusInternalServerError, fmt.Errorf("updated game for user: %w", err)
		}
	}

	return model.MapDbUserGameToUserGame(dbUserGame), http.StatusOK, nil
}
