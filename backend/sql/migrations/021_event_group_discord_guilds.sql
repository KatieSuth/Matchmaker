-- +goose Up
-- Host-selected Discord servers that lock an event group. Empty table for a group means unrestricted.
CREATE TABLE event_group_discord_guilds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event_group_id UUID NOT NULL REFERENCES event_groups(id) ON DELETE CASCADE,
    guild_id TEXT NOT NULL,
    guild_name TEXT NOT NULL,
    UNIQUE (event_group_id, guild_id)
);

-- +goose Down
DROP TABLE IF EXISTS event_group_discord_guilds;
