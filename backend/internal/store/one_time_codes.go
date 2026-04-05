package store

import (
	"context"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *PostgresStore) CreateOneTimeCode(ctx context.Context, code string, userID uuid.UUID) error {
	return s.q.CreateOneTimeCode(ctx, db.CreateOneTimeCodeParams{
		Code:   code,
		UserID: userID,
	})
}

func (s *PostgresStore) ConsumeOneTimeCode(ctx context.Context, code string) (uuid.UUID, error) {
	userID, err := s.q.ConsumeOneTimeCode(ctx, code)
	if err == pgx.ErrNoRows {
		return uuid.UUID{}, nil
	}
	return userID, err
}
