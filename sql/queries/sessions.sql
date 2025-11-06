-- name: CreateSession :one
INSERT INTO sessions (
name, user_id )
VALUES ( $1, $2)
RETURNING *;

-- name: GetUserSessions :many
SELECT * FROM sessions 
WHERE user_id = $1
ORDER BY created_at DESC;