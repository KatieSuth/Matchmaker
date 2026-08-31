-- name: DeleteEventGroupDiscordGuildsByGroupID :exec
DELETE FROM event_group_discord_guilds
WHERE event_group_id = $1;

-- name: InsertEventGroupDiscordGuild :one
INSERT INTO event_group_discord_guilds (
    id,
    event_group_id,
    guild_id,
    guild_name,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    NOW(),
    NOW()
)
RETURNING *;

-- name: ListEventGroupDiscordGuildsByGroupID :many
SELECT id, created_at, updated_at, event_group_id, guild_id, guild_name
FROM event_group_discord_guilds
WHERE event_group_id = $1
ORDER BY created_at ASC, id ASC;
