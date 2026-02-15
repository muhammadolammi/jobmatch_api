-- name: CreatePlan :one
INSERT INTO plans (
name,amount,currency,daily_limit, description  )
VALUES ( $1, $2, $3,$4,$5)
RETURNING *; 

-- name: GetPlans :many
SELECT * FROM plans;


-- name: GetPlanWithName :one
SELECT * FROM plans WHERE name=$1;
-- name: GetPlan :one
SELECT * FROM plans WHERE id=$1;

-- name: GetPlanWithPlaneCode :one
SELECT * FROM plans WHERE plan_code=$1;

-- name: UpdatePlanCode :exec
UPDATE plans
SET 
  plan_code = $1
WHERE id = $2;

-- name: UpdatePlanSubscriptionPage :exec
UPDATE plans
SET 
  subscription_page = $1
WHERE id = $2;





-- name: UpdatePLanSubscriptionPage :exec
UPDATE plans
SET 
  subscription_page = $1
WHERE id = $2;