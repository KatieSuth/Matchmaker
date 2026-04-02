-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expires_at, created_at, updated_at)
VALUES (
    $1,
    $2,
    $3,
    NOW(),
    NOW()
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: RevokeToken :exec
DELETE FROM refresh_tokens WHERE token = $1;
