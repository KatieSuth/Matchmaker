-- name: CreateEventGroup :one
INSERT INTO event_groups (id, owner_id, game_mode_id, sub_min, registration_open, region, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4,
    $5,
    NOW(),
    NOW()
)
RETURNING *;

-- name: CreateEvent :exec
INSERT INTO events (id, group_id, start_time, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    NOW(),
    NOW()
);

-- name: GetEventGroupById :one
SELECT * FROM event_groups WHERE id = $1;

-- name: GetEventsByGroupId :many
SELECT * FROM events WHERE group_id = $1 ORDER BY start_time ASC;

-- name: GetEventsForUser :many
WITH grouped AS (
    SELECT
        EG.id AS id,
        G.name AS game_name,
        GM.name AS game_mode,
        MIN(E.start_time)::TIMESTAMPTZ AS event_date,
        H.id AS host_id,
        COALESCE(H.discord_name, '') AS host_name,
        COUNT(DISTINCT RA.user_id)::INT AS registered_count,
        EG.registration_open
    FROM events AS E
    JOIN event_groups AS EG ON EG.id = E.group_id
    JOIN game_modes AS GM ON GM.id = EG.game_mode_id
    JOIN games AS G ON G.id = GM.game_id
    JOIN users AS H ON H.id = EG.owner_id
    LEFT JOIN registrations AS RA ON RA.event_id = E.id
    WHERE
        (
            (sqlc.arg(hosting)::BOOL = TRUE AND EG.owner_id = sqlc.arg(user_id))
            OR
            (sqlc.arg(hosting)::BOOL = FALSE AND EXISTS (
                SELECT 1
                FROM events AS ER
                JOIN registrations AS RM ON RM.event_id = ER.id
                WHERE ER.group_id = EG.id AND RM.user_id = sqlc.arg(user_id)
            ))
        )
        AND (NOT sqlc.arg(apply_past_filter)::BOOL OR (CASE WHEN sqlc.arg(past)::BOOL THEN E.start_time < sqlc.arg(boundary_time)::TIMESTAMPTZ ELSE E.start_time >= sqlc.arg(boundary_time)::TIMESTAMPTZ END))
        AND (NOT sqlc.arg(has_from)::BOOL OR E.start_time >= sqlc.arg(from_time)::TIMESTAMPTZ)
        AND (NOT sqlc.arg(has_to)::BOOL OR E.start_time < sqlc.arg(to_time)::TIMESTAMPTZ)
        AND (NOT sqlc.arg(has_game_id)::BOOL OR G.id = sqlc.arg(game_id)::UUID)
    GROUP BY EG.id, G.name, GM.name, H.id, H.discord_name, EG.registration_open
)
SELECT *
FROM grouped
WHERE (NOT sqlc.arg(has_cursor)::BOOL OR (event_date, id) > (sqlc.arg(cursor_time)::TIMESTAMPTZ, sqlc.arg(cursor_id)::UUID))
ORDER BY event_date ASC, id ASC
LIMIT sqlc.arg(limit_count)::INT;