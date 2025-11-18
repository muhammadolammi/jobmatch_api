-- +goose Up
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    provider TEXT NOT NULL,                -- e.g. "stripe"
    provider_payment_id TEXT NOT NULL,     -- stripe session or payment intent ID
    amount INTEGER NOT NULL,               -- stored in kobo/cents
    currency TEXT NOT NULL DEFAULT 'ngn',
    status TEXT NOT NULL,                  -- "pending", "success", "failed"
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_payments_users
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_payments_user_id ON payments(user_id);

-- +goose Down
DROP TABLE payments;
