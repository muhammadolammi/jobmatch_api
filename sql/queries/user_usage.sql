-- name: GetUserUsage :one
SELECT user_id, count, max_daily, last_used_at FROM user_usages WHERE user_id = $1;

-- name: InsertUserUsage :exec
INSERT INTO user_usages (user_id, count, max_daily, last_used_at)
VALUES ($1, $2, $3, $4);

-- name: UpdateUserUsage :exec
UPDATE user_usages
SET count = $2, last_used_at = $3
WHERE user_id = $1;