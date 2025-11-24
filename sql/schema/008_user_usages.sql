-- +goose Up
CREATE TABLE IF NOT EXISTS user_daily_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL,
    count INTEGER NOT NULL DEFAULT 0,
    max_daily INTEGER NOT NULL DEFAULT 2,    -- default free plan limit
    last_used_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user_daily_sages_users
      FOREIGN KEY (user_id)
      REFERENCES users(id)
      ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_daily_sages_users ON user_daily_usages(user_id);
-- +goose Down
DROP INDEX IF EXISTS idx_user_daily_sages_users;
DROP TABLE IF EXISTS user_daily_usages;
