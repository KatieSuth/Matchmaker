-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, discord_id, discord_name, image_url)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetUserByName :one
SELECT * FROM users WHERE discord_name = $1;

-- name: GetUserByDiscordID :one
SELECT * FROM users WHERE discord_id = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUser :one
UPDATE users
SET updated_at = NOW(),
    discord_id = $1,
    discord_name = $2,
    image_url = $3,
    pronouns = $4,
    show_pronouns = $5
WHERE id = $6
RETURNING *;