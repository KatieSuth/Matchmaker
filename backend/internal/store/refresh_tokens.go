package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) CreateNewRefreshToken(ctx context.Context, refreshTokenHash string, userID uuid.UUID, expires time.Time) (model.RefreshToken, error) {
	dbRefresh, err := s.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		Token:     refreshTokenHash,
		UserID:    userID,
		ExpiresAt: expires,
	})
	if err != nil {
		return model.RefreshToken{}, fmt.Errorf("creating refresh token for user %s: %w", userID.String(), err)
	}

	return model.MapDbRefreshTokenToRefreshToken(dbRefresh), nil
}

func (s *PostgresStore) GetRefreshToken(ctx context.Context, refreshTokenHash string) (model.RefreshToken, error) {
	dbRefresh, err := s.q.GetRefreshToken(ctx, refreshTokenHash)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return model.RefreshToken{}, fmt.Errorf("locating refresh token: %w", err)
	}

	return model.MapDbRefreshTokenToRefreshToken(dbRefresh), nil
}

func (s *PostgresStore) DeleteRefreshToken(ctx context.Context, refreshTokenHash string) error {
	err := s.q.RevokeToken(ctx, refreshTokenHash)
	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("deleting refresh token: %w", err)
	}

	return nil
}
