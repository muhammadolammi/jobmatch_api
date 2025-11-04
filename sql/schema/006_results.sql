-- +goose Up 
-- Enable the uuid-ossp extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    result jsonb  NOT NULL ,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    session_id UUID  NOT NULL,
     constraint fk_resumes_sessions
    foreign key (session_id) 
    REFERENCES sessions(id)
);

-- +goose Down
DROP TABLE results;