-- name: CreateOneTimeCode :exec
INSERT INTO one_time_codes (code, user_id)
VALUES (
    $1,
    $2
);

-- name: ConsumeOneTimeCode :one
DELETE FROM one_time_codes
WHERE code = $1 AND expires > now()
RETURNING user_id;
