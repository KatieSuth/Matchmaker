-- name: GetSystemGames :many
SELECT * FROM games WHERE owner_id IS NULL;