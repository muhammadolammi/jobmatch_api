-- +goose Up
CREATE TABLE user_professions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID  NOT NULL,
    profession_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user_professions_users
      FOREIGN KEY (user_id)
      REFERENCES users(id)
    ON DELETE CASCADE,
    CONSTRAINT fk_user_professions_professions
      FOREIGN KEY (profession_id)
      REFERENCES professions(id)
      ON DELETE CASCADE
);

-- +goose Down
DROP TABLE user_professions;