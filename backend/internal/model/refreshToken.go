package model

import (
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

type RefreshToken struct {
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

/* Mapping function to remove sensitive fields (not important
 * now since we just use Discord but could be later when working
 * with email/password authentication) and enable json mappings
 * (sqlc can do this but handling it manually gives more flexibility)
 */
func MapDbRefreshTokenToRefreshToken(dbRefresh db.RefreshToken) RefreshToken {
	return RefreshToken{
		Token:     dbRefresh.Token,
		UserID:    dbRefresh.UserID,
		ExpiresAt: dbRefresh.ExpiresAt,
		CreatedAt: dbRefresh.CreatedAt,
		UpdatedAt: dbRefresh.CreatedAt,
	}
}
