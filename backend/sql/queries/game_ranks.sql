-- name: GetRanksForGame :many
SELECT * FROM game_ranks WHERE game_id = $1;

-- name: GetRankById :one
SELECT * FROM game_ranks WHERE id = $1;