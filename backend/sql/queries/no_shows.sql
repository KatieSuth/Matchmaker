-- name: UpsertNoShow :exec
-- ON CONFLICT DO NOTHING keeps the original marked_absent_at when a player is re-marked.
INSERT INTO no_shows (event_id, user_id, host_id)
VALUES ($1, $2, $3)
ON CONFLICT (event_id, user_id) DO NOTHING;

-- name: DeleteNoShow :exec
DELETE FROM no_shows
WHERE event_id = $1 AND user_id = $2;

-- name: ListNoShowUserIdsForEvent :many
SELECT user_id
FROM no_shows
WHERE event_id = $1
ORDER BY user_id;
