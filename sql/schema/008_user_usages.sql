-- +goose Up
CREATE TABLE IF NOT EXISTS user_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL,
    count INTEGER NOT NULL DEFAULT 0,
    max_daily INTEGER NOT NULL DEFAULT 2,    -- default free plan limit
    last_used_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user_usages_users
      FOREIGN KEY (user_id)
      REFERENCES users(id)
      ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_usages_user_id ON user_usages(user_id);
-- +goose Down
DROP TABLE IF EXISTS user_usages;
