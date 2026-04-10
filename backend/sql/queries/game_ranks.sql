-- name: GetRanksForGame :many
SELECT * FROM game_ranks WHERE game_id = $1;