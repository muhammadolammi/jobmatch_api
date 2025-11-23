-- name: CheckSubscriptionExist :one
SELECT EXISTS(
    SELECT 1 FROM subscriptions WHERE user_id = $1
    );

-- name: CreateSubscription :one
INSERT INTO subscriptions (
 user_id, plan_id )
VALUES ( $1, $2)
RETURNING *;

-- name: GetSubscriptionWithUserID :one

SELECT * FROM subscriptions 
WHERE user_id=$1;


-- name: UpdateSubscriptionPlan :exec
UPDATE subscriptions
SET 
  plan_id = $1
WHERE id = $2;
