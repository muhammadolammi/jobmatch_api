-- name: CheckSubscriptionExist :one
SELECT EXISTS(
    SELECT 1 FROM subscriptions WHERE user_id = $1
    );