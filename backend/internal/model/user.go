package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	DiscordID    *string   `json:"discord_id"`
	DiscordName  *string   `json:"discord_name"`
	ImageUrl     *string   `json:"image_url"`
	Pronouns     *string   `json:"pronouns"`
	ShowPronouns bool      `json:"show_pronouns"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Region       *string   `json:"region"`
}
