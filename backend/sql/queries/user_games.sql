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

-- name: GetGameForUserByIds :one
SELECT *
FROM user_games AS UG
WHERE user_id = $1 AND game_id = $2;

-- name: CreateGameForUser :one
INSERT INTO user_games (user_id, game_id, created_at, updated_at, in_game_name, current_rank, peak_rank, show_rank)
VALUES (
    $1,
    $2,
    NOW(),
    NOW(),
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: UpdateGameForUser :one
UPDATE user_games
SET updated_at = NOW(),
    in_game_name = $1,
    current_rank = $2,
    peak_rank = $3,
    show_rank = $4
WHERE user_id = $5
AND game_id = $6
RETURNING *;