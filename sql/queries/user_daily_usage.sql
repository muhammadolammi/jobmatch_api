-- name: GetUserUsage :one
SELECT user_id, count, max_daily, last_used_at FROM user_daily_usages WHERE user_id = $1;

-- name: InsertUserUsage :exec
INSERT INTO user_daily_usages (user_id, count, max_daily, last_used_at)
VALUES ($1, $2, $3, $4);

-- name: UpdateUserUsage :exec
UPDATE user_daily_usages
SET count = $2, last_used_at = $3
WHERE user_id = $1;


-- name: UpdateUserDailyUsageLimit :exec
INSERT INTO user_daily_usages (user_id, max_daily, count, last_used_at)
VALUES ($1, $2, 0, NOW())
ON CONFLICT (user_id)
DO UPDATE SET max_daily = $2;