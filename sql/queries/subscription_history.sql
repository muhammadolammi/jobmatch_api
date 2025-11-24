-- name: CreateSubscriptionHistory :one
INSERT INTO subscription_history (
 user_id, subscription_id,event_type, old_value, new_value, event_source )
VALUES ( $1, $2,$3,$4,$5,$6)
RETURNING *;