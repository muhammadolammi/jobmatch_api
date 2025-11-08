-- +goose Up
-- Allow multiple sessions per user
ALTER TABLE sessions
DROP CONSTRAINT IF EXISTS sessions_user_id_key;

-- +goose Down
-- Revert: make user_id unique again (old behavior)
ALTER TABLE sessions
ADD CONSTRAINT sessions_user_id_key UNIQUE (user_id);
