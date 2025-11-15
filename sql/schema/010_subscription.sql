-- +goose Up
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',     -- pending, active, canceled, expired
    expires_at TIMESTAMP ,
    canceled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    next_payment_date TIMESTAMP NOT NULL ,
    subscription_code  TEXT UNIQUE ,
    plan_id UUID NOT NULL,
    CONSTRAINT fk_subscriptions_users
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_subscriptions_plans
    FOREIGN KEY (plan_id)
    REFERENCES plans(id)
    ON DELETE CASCADE
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);

-- +goose Down
DROP TABLE subscriptions;
