package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidGame        = errors.New("invalid game")
	ErrCurrentRankMissing = errors.New("current rank must not be empty")
	ErrPeakRankMissing    = errors.New("peak rank must not be empty")
	ErrInvalidCurrentRank = errors.New("invalid current rank")
	ErrInvalidPeakRank    = errors.New("invalid peak rank")
	ErrInvalidRankOrder   = errors.New("peak rank must be greater than or equal to current rank")
)

func (s *PostgresStore) GetUserGamesForUser(ctx context.Context, userID uuid.UUID) ([]model.UserGame, error) {
	dbUserGames, err := s.q.GetGamesForUser(ctx, userID)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return []model.UserGame{}, fmt.Errorf("looking up user's games: %w", err)
	}

	return model.MapDbUserGamesToUserGames(dbUserGames), nil
}

// UpsertGameForUser validates rank ordering and game/rank foreign keys, then inserts or updates
// the user_games row.
func (s *PostgresStore) UpsertGameForUser(ctx context.Context, userID uuid.UUID, ug model.UserGame) (model.UserGame, error) {
	/*** validate provided info ***/

	//make sure game exists
	_, err := s.q.GetGameById(ctx, ug.GameID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserGame{}, fmt.Errorf("%w: %s", ErrInvalidGame, ug.GameID.String())
		}
		return model.UserGame{}, fmt.Errorf("looking up game by ID: %w", err)
	}

	//make sure we got existing ranks
	if ug.CurrentRank == nil || *ug.CurrentRank == uuid.Nil {
		return model.UserGame{}, ErrCurrentRankMissing
	}

	if ug.PeakRank == nil || *ug.PeakRank == uuid.Nil {
		return model.UserGame{}, ErrPeakRankMissing
	}

	currentRank, err := s.q.GetRankById(ctx, *ug.CurrentRank)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return model.UserGame{}, fmt.Errorf("looking up current rank by ID: %w", err)
	} else if err != nil {
		return model.UserGame{}, fmt.Errorf("%w: %s", ErrInvalidCurrentRank, ug.CurrentRank.String())
	}

	peakRank, err := s.q.GetRankById(ctx, *ug.PeakRank)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return model.UserGame{}, fmt.Errorf("looking up peak rank by ID: %w", err)
	} else if err != nil {
		return model.UserGame{}, fmt.Errorf("%w: %s", ErrInvalidPeakRank, ug.PeakRank.String())
	}

	//make sure ranks are logical
	if currentRank.Order > peakRank.Order {
		return model.UserGame{}, ErrInvalidRankOrder
	}

	avgRankID, _, err := s.resolveAvgRankID(ctx, ug.GameID, currentRank.Order, peakRank.Order)
	if err != nil {
		return model.UserGame{}, err
	}

	/*** Upsert ***/
	_, err = s.q.GetGameForUserByIds(ctx, db.GetGameForUserByIdsParams{
		UserID: userID,
		GameID: ug.GameID,
	})

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		// Genuine DB error — surface it
		return model.UserGame{}, fmt.Errorf("looking up user game: %w", err)
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
			AvgRank:     &avgRankID,
			ShowRank:    ug.ShowRank,
		})
		if err != nil {
			return model.UserGame{}, fmt.Errorf("creating game for user: %w", err)
		}
	} else {
		//userGame link exists, update it
		dbUserGame, err = s.q.UpdateGameForUser(ctx, db.UpdateGameForUserParams{
			InGameName:  *ug.InGameName,
			CurrentRank: ug.CurrentRank,
			PeakRank:    ug.PeakRank,
			AvgRank:     &avgRankID,
			ShowRank:    ug.ShowRank,
			UserID:      userID,
			GameID:      ug.GameID,
		})

		if err != nil {
			return model.UserGame{}, fmt.Errorf("updated game for user: %w", err)
		}
	}

	return model.MapDbUserGameToUserGame(dbUserGame), nil
}

func (s *PostgresStore) DeleteGameForUser(ctx context.Context, userID, gameID uuid.UUID) error {
	if err := s.q.DeleteGameForUser(ctx, db.DeleteGameForUserParams{
		UserID: userID,
		GameID: gameID,
	}); err != nil {
		return fmt.Errorf("deleting game for user: %w", err)
	}
	return nil
}
