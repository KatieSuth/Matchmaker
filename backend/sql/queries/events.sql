-- name: CreateEventGroup :one
INSERT INTO event_groups (id, owner_id, sub_min, registration_open, region, sort_logic, created_at, updated_at)
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
INSERT INTO events (id, group_id, start_time, game_mode_id, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
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
    CAST(
        CASE
            WHEN U.show_pronouns = TRUE THEN COALESCE(U.pronouns, '')
            ELSE ''
        END
        AS TEXT
    ) AS owner_pronouns,
    CAST(COALESCE(
        (SELECT STRING_AGG(DISTINCT gm_agg.name, ', ' ORDER BY gm_agg.name)
         FROM events E_agg
         JOIN game_modes gm_agg ON gm_agg.id = E_agg.game_mode_id
         WHERE E_agg.group_id = EG.id),
        ''
    ) AS VARCHAR(4096)) AS game_mode_name,
    CAST(
        CASE
            WHEN (SELECT COUNT(DISTINCT E_mid.game_mode_id) FROM events E_mid WHERE E_mid.group_id = EG.id) = 1
            THEN (SELECT E_mid.game_mode_id FROM events E_mid WHERE E_mid.group_id = EG.id LIMIT 1)
            ELSE '00000000-0000-0000-0000-000000000000'::uuid
        END
        AS UUID
    ) AS game_mode_id,
    CAST(
        COALESCE(
            (SELECT G_first.id
             FROM events E_first
             JOIN game_modes GM_first ON GM_first.id = E_first.game_mode_id
             JOIN games G_first ON G_first.id = GM_first.game_id
             WHERE E_first.group_id = EG.id
             ORDER BY E_first.start_time ASC NULLS LAST
             LIMIT 1),
            '00000000-0000-0000-0000-000000000000'::uuid
        )
        AS UUID
    ) AS game_id,
    CAST(
        COALESCE(
            (SELECT G_first.name
             FROM events E_first
             JOIN game_modes GM_first ON GM_first.id = E_first.game_mode_id
             JOIN games G_first ON G_first.id = GM_first.game_id
             WHERE E_first.group_id = EG.id
             ORDER BY E_first.start_time ASC NULLS LAST
             LIMIT 1),
            ''
        )
        AS TEXT
    ) AS game_name,
    CAST(
        CASE
            WHEN (SELECT COUNT(DISTINCT GM_sz.team_size) FROM events E_sz JOIN game_modes GM_sz ON GM_sz.id = E_sz.game_mode_id WHERE E_sz.group_id = EG.id) = 1
            THEN (SELECT MIN(GM_sz.team_size) FROM events E_sz JOIN game_modes GM_sz ON GM_sz.id = E_sz.game_mode_id WHERE E_sz.group_id = EG.id)
            ELSE 0
        END
        AS INT
    ) AS team_size,
    EG.sub_min,
    EG.registration_open,
    EG.region,
    EG.sort_logic,
    EG.created_at,
    EG.updated_at
FROM event_groups AS EG
JOIN users AS U ON U.id = EG.owner_id
WHERE EG.id = $1;

-- name: GetGroupEventsSummary :many
SELECT
    E.id,
    E.start_time,
    E.game_mode_id,
    GM.name AS game_mode_name,
    GM.team_size,
    COUNT(DISTINCT R.user_id)::INT AS registered_count,
    COUNT(DISTINCT L.id)::INT AS lobbies_count,
    COALESCE(BOOL_OR(R.user_id = sqlc.arg(viewer_id)), FALSE)::BOOL AS player_registered
FROM events AS E
JOIN game_modes AS GM ON GM.id = E.game_mode_id
LEFT JOIN registrations AS R ON R.event_id = E.id
LEFT JOIN lobbies AS L ON L.event_id = E.id
WHERE E.group_id = $1
GROUP BY E.id, E.start_time, E.game_mode_id, GM.name, GM.team_size
ORDER BY E.start_time ASC;

-- name: GetEventWithGroupById :one
SELECT
    E.id,
    E.group_id,
    G.id AS game_id,
    EG.owner_id,
    EG.registration_open
FROM events AS E
JOIN event_groups AS EG ON EG.id = E.group_id
JOIN game_modes AS GM ON GM.id = E.game_mode_id
JOIN games AS G ON G.id = GM.game_id
WHERE E.id = $1;

-- name: CountLobbiesByGroupId :one
SELECT COUNT(*)::INT
FROM lobbies AS L
JOIN events AS E ON E.id = L.event_id
WHERE E.group_id = $1;

-- name: UpdateEventGroupSettings :one
UPDATE event_groups
SET region = $2, sub_min = $3, sort_logic = $4, registration_open = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateEventSchedule :execrows
UPDATE events
SET start_time = $2, game_mode_id = $3, updated_at = NOW()
WHERE id = $1 AND group_id = $4;

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
        CAST(STRING_AGG(DISTINCT GM.name, ', ' ORDER BY GM.name) AS VARCHAR(4096)) AS game_mode,
        MIN(E.start_time)::TIMESTAMPTZ AS event_date,
        H.id AS host_id,
        COALESCE(H.discord_name, '') AS host_name,
        COUNT(DISTINCT RA.user_id)::INT AS registered_count,
        EG.registration_open
    FROM events AS E
    JOIN event_groups AS EG ON EG.id = E.group_id
    JOIN game_modes AS GM ON GM.id = E.game_mode_id
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
    GROUP BY EG.id, G.name, H.id, H.discord_name, EG.registration_open
)
SELECT *
FROM grouped
WHERE (NOT sqlc.arg(has_cursor)::BOOL OR (event_date, id) > (sqlc.arg(cursor_time)::TIMESTAMPTZ, sqlc.arg(cursor_id)::UUID))
ORDER BY event_date ASC, id ASC
LIMIT sqlc.arg(limit_count)::INT;
