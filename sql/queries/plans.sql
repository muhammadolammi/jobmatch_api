-- name: CreatePlan :one
INSERT INTO plans (
name, plan_code,amount,currency,daily_limit, description  )
VALUES ( $1, $2, $3,$4,$5,$6 )
RETURNING *; 

-- name: GetPlans :many
SELECT * FROM plans;