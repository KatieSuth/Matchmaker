-- name: GetSystemGames :many
SELECT * FROM games WHERE owner_id IS NULL;

-- name: GetGameById :one
SELECT * FROM games WHERE id = $1;