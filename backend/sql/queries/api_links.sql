-- name: UpsertApiLink :one
INSERT INTO api_links (id, user_id, name, refresh_token, refresh_token_iv, key_id, created_at, updated_at)
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
ON CONFLICT (user_id, name)
DO UPDATE SET
    refresh_token = EXCLUDED.refresh_token,
    refresh_token_iv = EXCLUDED.refresh_token_iv,
    key_id = EXCLUDED.key_id,
    updated_at = NOW()
RETURNING *;

-- name: GetApiLinkByUserAndName :one
SELECT * FROM api_links
WHERE user_id = $1 AND name = $2;

-- name: GetApiLinkByUserAndNameForUpdate :one
SELECT * FROM api_links
WHERE user_id = $1 AND name = $2
FOR UPDATE;

-- name: DeleteApiLinkByUserAndName :exec
DELETE FROM api_links
WHERE user_id = $1 AND name = $2;
