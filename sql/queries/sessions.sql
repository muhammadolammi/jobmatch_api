-- name: CreateSession :one
INSERT INTO sessions (
name, user_id, job_title, job_description )
VALUES ( $1, $2, $3,$4)
RETURNING *;

-- name: GetUserSessions :many
SELECT * FROM sessions 
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetSession :one 
SELECT * FROM sessions 
WHERE id = $1;

-- name: UpdateSessionStatus :exec
UPDATE sessions 
SET status=$1
WHERE id=$2;
