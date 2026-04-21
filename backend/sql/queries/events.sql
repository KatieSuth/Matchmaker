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