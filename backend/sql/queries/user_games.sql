-- name: GetGamesForUser :many
SELECT UG.user_id, UG.game_id, UG.in_game_name, UG.current_rank, UG.peak_rank, UG.show_rank, UG.created_at, UG.updated_at,
       G.name as game_name,
       GR_Current.name AS current_rank_name,
       GR_Peak.name AS peak_rank_name
FROM user_games AS UG
JOIN games AS G ON (UG.game_id = G.id)
JOIN game_ranks AS GR_Current ON (UG.current_rank = GR_Current.id)
JOIN game_ranks AS GR_Peak ON (UG.peak_rank = GR_Peak.id)
WHERE UG.user_id = $1;