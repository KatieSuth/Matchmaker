-- name: GetSystemGames :many
SELECT * FROM games WHERE owner_id IS NULL;

-- name: GetGameById :one
SELECT * FROM games WHERE id = $1;

-- name: GetUserGames :many
SELECT * FROM games WHERE owner_id IS NULL OR owner_id = $1;

-- name: GetGameModes :many
SELECT * FROM game_modes WHERE game_id = $1;

-- name: GetGameModeById :one
SELECT * FROM game_modes WHERE id = $1;