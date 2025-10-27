-- +goose Up 
-- Enable the uuid-ossp extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    result jsonb  NOT NULL ,
    session_id TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp


    
);

-- +goose Down
DROP TABLE results;