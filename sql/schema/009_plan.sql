-- +goose Up
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,        -- e.g. "free", "pro", "team"
    plan_code TEXT UNIQUE ,
    amount INTEGER NOT NULL,           -- smallest currency unit (kobo, cents)
    currency TEXT NOT NULL ,           -- NGN, USD, EUR
    interval TEXT NOT NULL DEFAULT 'monthly',           -- "daily", "monthly", "yearly"
    daily_limit INTEGER NOT NULL,     -- analyses per day
    description TEXT,                 -- optional (for UI/marketing)
    subscription_page TEXT UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE plans;
