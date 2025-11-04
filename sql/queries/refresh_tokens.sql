-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
user_id, expires_at,
token  )
VALUES ( $1, $2, $3)
RETURNING *;


-- name: UpdateRefreshToken :exec
UPDATE refresh_tokens
SET 
  token = $1,
  expires_at = $2
WHERE user_id = $3;


-- name: RefreshTokenExists :one
SELECT EXISTS (
    SELECT 1
    FROM refresh_tokens
    WHERE token = $1
);
