package model

import (
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

// ApiLink is a per-provider encrypted refresh token. Ciphertext fields are omitted from JSON.
type ApiLink struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Name           string    `json:"name"`
	RefreshToken   string    `json:"-"`
	RefreshTokenIv string    `json:"-"`
	KeyID          string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MapDbApiLinkToApiLink copies a sqlc api_links row into the API model.
func MapDbApiLinkToApiLink(dbLink db.ApiLink) ApiLink {
	return ApiLink{
		ID:             dbLink.ID,
		UserID:         dbLink.UserID,
		Name:           dbLink.Name,
		RefreshToken:   dbLink.RefreshToken,
		RefreshTokenIv: dbLink.RefreshTokenIv,
		KeyID:          dbLink.KeyID,
		CreatedAt:      dbLink.CreatedAt,
		UpdatedAt:      dbLink.UpdatedAt,
	}
}
