-- +goose Up
ALTER TABLE sessions DROP CONSTRAINT sessions_name_key;

ALTER TABLE sessions
ADD CONSTRAINT unique_user_session_name UNIQUE (user_id, name);

-- +goose Down
ALTER TABLE sessions DROP CONSTRAINT unique_user_session_name;

ALTER TABLE sessions
ADD CONSTRAINT sessions_name_key UNIQUE (name);