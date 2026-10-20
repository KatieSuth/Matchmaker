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

-- name: CountRecentNoShowsForHost :many
-- Per-player no-show counts at one host's games whose start_time falls in [since, until].
-- Counted by game date, not when the mark was recorded.
SELECT ns.user_id, COUNT(*)::INT AS no_show_count
FROM no_shows ns
JOIN events e ON e.id = ns.event_id
WHERE ns.host_id = sqlc.arg(host_id)
  AND ns.user_id = ANY(sqlc.arg(user_ids)::uuid[])
  AND e.start_time BETWEEN sqlc.arg(since) AND sqlc.arg(until)
GROUP BY ns.user_id;

-- name: ListRecentNoShowDatesForHostUser :many
-- Start times of one host's games in [since, until] where this user was marked absent, oldest first.
SELECT e.start_time
FROM no_shows ns
JOIN events e ON e.id = ns.event_id
WHERE ns.host_id = sqlc.arg(host_id)
  AND ns.user_id = sqlc.arg(user_id)
  AND e.start_time BETWEEN sqlc.arg(since) AND sqlc.arg(until)
ORDER BY e.start_time ASC;
