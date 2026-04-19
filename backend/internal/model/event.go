package model

import (
	"time"

	"github.com/google/uuid"
)

type DashboardEvent struct {
	ID               uuid.UUID `json:"id"`
	GameName         string    `json:"game_name"`
	GameMode         string    `json:"game_mode"`
	EventDate        time.Time `json:"event_date"`
	HostID           uuid.UUID `json:"host_id"`
	HostName         string    `json:"host_name"`
	RegisteredCount  int       `json:"registered_count"`
	RegistrationOpen bool      `json:"registration_open"`
}
