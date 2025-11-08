-- +goose Up

CREATE TABLE user_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    count INTEGER NOT NULL DEFAULT 0,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    max_daily INTEGER NOT NULL,
    user_id UUID UNIQUE NOT NULL,
   constraint fk_employer_profiles_users
    foreign key (user_id) 
    REFERENCES users(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE user_usages;