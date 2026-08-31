package model

import (
	"github.com/KatieSuth/MatchmakerAPI/internal/db"
)

// DiscordGuild is a Discord server id and display name used in APIs and event locks.
type DiscordGuild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MapDbEventGroupDiscordGuildToDiscordGuild copies a junction row into the API guild shape.
func MapDbEventGroupDiscordGuildToDiscordGuild(row db.EventGroupDiscordGuild) DiscordGuild {
	return DiscordGuild{
		ID:   row.GuildID,
		Name: row.GuildName,
	}
}

// MapDbEventGroupDiscordGuildsToDiscordGuilds maps junction rows; nil input becomes an empty slice.
func MapDbEventGroupDiscordGuildsToDiscordGuilds(rows []db.EventGroupDiscordGuild) []DiscordGuild {
	out := make([]DiscordGuild, 0, len(rows))
	for _, row := range rows {
		out = append(out, MapDbEventGroupDiscordGuildToDiscordGuild(row))
	}
	return out
}
