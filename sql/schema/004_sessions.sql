-- +goose Up 
-- Enable the uuid-ossp extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT UNIQUE NOT NULL,
    user_id UUID UNIQUE NOT NULL,
    constraint fk_sessions_users
    foreign key (user_id) 
    REFERENCES users(id)
);

-- +goose Down
DROP TABLE sessions;

