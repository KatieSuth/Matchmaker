-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, discord_id, discord_name, image_url, display_name)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3,
    $4
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
    display_name = $1,
    pronouns = $2,
    show_pronouns = $3,
    region = $4,
    new_user = false
WHERE id = $5
RETURNING *;

-- name: UpdateUserFromLogin :one
UPDATE users
SET updated_at = NOW(),
    discord_name = $1,
    image_url = $2
WHERE id = $3
RETURNING *;
