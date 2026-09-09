-- +goose Up
CREATE TABLE job_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employer_id UUID  NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_employer_job_posts_employer
      FOREIGN KEY (employer_id)
      REFERENCES employer_profiles(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE job_posts;