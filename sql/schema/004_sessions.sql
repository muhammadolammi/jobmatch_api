-- +goose Up 
-- Enable the uuid-ossp extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT  NOT NULL,
    user_id UUID  NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    job_title TEXT NOT NULL ,
    job_description TEXT NOT NULL ,
    CONSTRAINT unique_user_session_name UNIQUE (user_id, name),
    constraint fk_sessions_users
    foreign key (user_id) 
    REFERENCES users(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE sessions;

