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


-- name: UpdateSubscriptionStatus :exec
UPDATE subscriptions
SET 
  status = $1
WHERE id = $2;


-- name: UpdateSubscriptionNextPaymentDate :exec
UPDATE subscriptions
SET 
  next_payment_date = $1
WHERE id = $2;


-- name: UpdateSubscriptionForActivation :exec
UPDATE subscriptions
SET 
  next_payment_date = $1,
  status = $2,
  paystack_sub_code = $3

WHERE id = $4;