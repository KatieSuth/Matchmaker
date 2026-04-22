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

-- name: GetEventGroupDetailById :one
SELECT
    EG.id,
    EG.owner_id,
    COALESCE(U.discord_name, '') AS owner_name,
    EG.game_mode_id,
    GM.name AS game_mode_name,
    G.id AS game_id,
    G.name AS game_name,
    GM.team_size,
    EG.sub_min,
    EG.registration_open,
    EG.region,
    EG.created_at,
    EG.updated_at
FROM event_groups AS EG
JOIN users AS U ON U.id = EG.owner_id
JOIN game_modes AS GM ON GM.id = EG.game_mode_id
JOIN games AS G ON G.id = GM.game_id
WHERE EG.id = $1;

-- name: GetGroupEventsSummary :many
SELECT
    E.id,
    E.start_time,
    COUNT(DISTINCT R.user_id)::INT AS registered_count,
    COUNT(DISTINCT L.id)::INT AS lobbies_count,
    COALESCE(BOOL_OR(R.user_id = sqlc.arg(viewer_id)), FALSE)::BOOL AS player_registered
FROM events AS E
LEFT JOIN registrations AS R ON R.event_id = E.id
LEFT JOIN lobbies AS L ON L.event_id = E.id
WHERE E.group_id = $1
GROUP BY E.id
ORDER BY E.start_time ASC;

-- name: GetEventWithGroupById :one
SELECT
    E.id,
    E.group_id,
    EG.owner_id,
    EG.registration_open
FROM events AS E
JOIN event_groups AS EG ON EG.id = E.group_id
WHERE E.id = $1;

-- name: CountLobbiesByGroupId :one
SELECT COUNT(*)::INT
FROM lobbies AS L
JOIN events AS E ON E.id = L.event_id
WHERE E.group_id = $1;

-- name: UpdateEventGroupSettings :one
UPDATE event_groups
SET region = $2, sub_min = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetEventGroupRegistrationOpen :one
UPDATE event_groups
SET registration_open = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEventGroupById :execrows
DELETE FROM event_groups
WHERE id = $1;

-- name: CreateLobby :one
INSERT INTO lobbies (id, event_id, host, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, NOW(), NOW())
RETURNING *;

-- name: CreatePlayer :exec
INSERT INTO players (lobby_id, user_id, team_number, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW());

-- name: DeletePlayersByGroupId :exec
DELETE FROM players
WHERE lobby_id IN (
    SELECT L.id
    FROM lobbies AS L
    JOIN events AS E ON E.id = L.event_id
    WHERE E.group_id = $1
);

-- name: DeleteLobbiesByGroupId :exec
DELETE FROM lobbies
WHERE event_id IN (
    SELECT id
    FROM events
    WHERE group_id = $1
);

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