package store

import (
	"context"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
)

// UpsertApiLink inserts or replaces the encrypted refresh token for this user and provider name.
func (s *PostgresStore) UpsertApiLink(ctx context.Context, userID uuid.UUID, name, ciphertext, nonce, keyID string) (model.ApiLink, error) {
	dbLink, err := s.q.UpsertApiLink(ctx, db.UpsertApiLinkParams{
		UserID:         userID,
		Name:           name,
		RefreshToken:   ciphertext,
		RefreshTokenIv: nonce,
		KeyID:          keyID,
	})
	if err != nil {
		return model.ApiLink{}, fmt.Errorf("upserting api link for user %s name %s: %w", userID.String(), name, err)
	}

	return model.MapDbApiLinkToApiLink(dbLink), nil
}

// GetApiLinkByUserAndName loads the encrypted refresh token row for this user and provider name.
func (s *PostgresStore) GetApiLinkByUserAndName(ctx context.Context, userID uuid.UUID, name string) (model.ApiLink, error) {
	dbLink, err := s.q.GetApiLinkByUserAndName(ctx, db.GetApiLinkByUserAndNameParams{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		return model.ApiLink{}, fmt.Errorf("locating api link for user %s name %s: %w", userID.String(), name, err)
	}

	return model.MapDbApiLinkToApiLink(dbLink), nil
}

// GetApiLinkByUserAndNameForUpdate loads the encrypted blob and row-locks it until the surrounding tx ends.
func (s *PostgresStore) GetApiLinkByUserAndNameForUpdate(ctx context.Context, userID uuid.UUID, name string) (model.ApiLink, error) {
	dbLink, err := s.q.GetApiLinkByUserAndNameForUpdate(ctx, db.GetApiLinkByUserAndNameForUpdateParams{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		return model.ApiLink{}, fmt.Errorf("locating api link for update for user %s name %s: %w", userID.String(), name, err)
	}

	return model.MapDbApiLinkToApiLink(dbLink), nil
}

// DeleteApiLinkByUserAndName removes the encrypted blob for this user and provider name.
func (s *PostgresStore) DeleteApiLinkByUserAndName(ctx context.Context, userID uuid.UUID, name string) error {
	err := s.q.DeleteApiLinkByUserAndName(ctx, db.DeleteApiLinkByUserAndNameParams{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		return fmt.Errorf("deleting api link for user %s name %s: %w", userID.String(), name, err)
	}
	return nil
}
