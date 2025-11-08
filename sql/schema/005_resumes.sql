-- +goose Up 
-- Enable the uuid-ossp extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE resumes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    original_filename TEXT  NOT NULL ,
    mime TEXT  NOT NULL ,
    size_bytes BIGINT NOT NULL,
    storage_provider TEXT NOT NULL,
    object_key TEXT NOT NULL,
    storage_url TEXT NOT NULL,
    upload_status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    session_id UUID  NOT NULL,
    constraint fk_resumes_sessions
    foreign key (session_id) 
    REFERENCES sessions(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE resumes;