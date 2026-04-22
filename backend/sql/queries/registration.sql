-- name: GetRegistrationDataByEventId :many
SELECT R.*, 
    U.discord_name, 
    CASE 
		WHEN sqlc.arg(viewer_is_host)::BOOL = TRUE OR U.show_pronouns = true THEN COALESCE(U.pronouns, '')
		ELSE ''
	END AS pronouns,
    CASE
        WHEN sqlc.arg(viewer_is_host)::BOOL = TRUE OR COALESCE(UG.show_rank, FALSE) = TRUE THEN COALESCE(GR.name, '')
        ELSE ''
    END AS current_rank_name
FROM registrations AS R 
JOIN users AS U ON R.user_id = U.id 
JOIN events AS E ON R.event_id = E.id
JOIN event_groups AS EG ON E.group_id = EG.id
JOIN game_modes AS GM ON GM.id = EG.game_mode_id
LEFT JOIN user_games AS UG ON R.user_id = UG.user_id AND UG.game_id = GM.game_id
LEFT JOIN game_ranks AS GR ON UG.current_rank = GR.id
WHERE event_id = $1
ORDER BY U.discord_name ASC;

-- name: GetRegistrationByEventAndUser :one
SELECT *
FROM registrations
WHERE event_id = $1 AND user_id = $2;

-- name: CreateRegistration :one
INSERT INTO registrations (event_id, user_id, can_substitute, can_lobby_host, duo_request, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING *;

-- name: UpdateRegistration :one
UPDATE registrations
SET can_substitute = $3, can_lobby_host = $4, duo_request = $5, updated_at = NOW()
WHERE event_id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteRegistration :exec
DELETE FROM registrations
WHERE event_id = $1 AND user_id = $2;

-- name: GetRegistrationsForEvent :many
SELECT user_id, can_lobby_host, created_at
FROM registrations
WHERE event_id = $1
ORDER BY created_at ASC;