package model

import (
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	DiscordID    *string   `json:"discord_id"`
	DiscordName  *string   `json:"discord_name"`
	ImageUrl     *string   `json:"image_url"`
	DisplayName  *string   `json:"display_name"`
	Pronouns     *string   `json:"pronouns"`
	ShowPronouns bool      `json:"show_pronouns"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Region       *string   `json:"region"`
	NewUser      bool      `json:"new_user"`
}

/* Mapping function to remove sensitive fields (not important
 * now since we just use Discord but could be later when working
 * with email/password authentication) and enable json mappings
 * (sqlc can do this but handling it manually gives more flexibility)
 */
func MapDbUserToUser(dbUser db.User) User {
	return User{
		ID:           dbUser.ID,
		DiscordID:    dbUser.DiscordID,
		DiscordName:  dbUser.DiscordName,
		ImageUrl:     dbUser.ImageUrl,
		DisplayName:  dbUser.DisplayName,
		Pronouns:     dbUser.Pronouns,
		ShowPronouns: dbUser.ShowPronouns,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Region:       dbUser.Region,
		NewUser:      dbUser.NewUser,
	}
}
