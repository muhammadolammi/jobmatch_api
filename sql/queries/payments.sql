-- -- name: CreatePayment :one
-- INSERT INTO payments (provider, provider_payment_id,  amount, currency, status, user_id)
-- VALUES ($1, $2, $3, $4, $5, $6)
-- RETURNING *;

