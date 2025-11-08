-- name: GetUserUsage :one
SELECT * FROM user_usages WHERE user_id = $1 LIMIT 1;

-- name: InsertUserUsage :exec
INSERT INTO user_usages (user_id, count, last_used_at, max_daily)
VALUES ($1, 1, NOW(), $2)
ON CONFLICT (user_id) DO UPDATE
SET count = user_usages.count + 1, last_used_at = NOW();

-- name: UpdateUserUsageCount :exec
UPDATE user_usages
SET count = $2, last_used_at = NOW()
WHERE user_id = $1;

-- name: UpdateUserMaxDaily :exec
UPDATE user_usages
SET max_daily = $1
WHERE user_id = $2;