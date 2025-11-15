-- +goose Up
CREATE TABLE subscription_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL,
    user_id UUID NOT NULL,
    event_type TEXT NOT NULL,     -- created, renewed, upgraded, downgraded, canceled, expired, payment_failed
    old_value TEXT,               -- previous plan or status
    new_value TEXT,               -- new plan or status
    event_source TEXT NOT NULL,   -- system, user, admin, webhook
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- payload jsonb NOT NULL,
    CONSTRAINT fk_history_subscription
        FOREIGN KEY (subscription_id)
        REFERENCES subscriptions(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_history_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_history_subscription_id ON subscription_history(subscription_id);
CREATE INDEX idx_history_user_id ON subscription_history(user_id);
CREATE INDEX idx_history_event_type ON subscription_history(event_type);

-- +goose Down
DROP TABLE subscription_history;
