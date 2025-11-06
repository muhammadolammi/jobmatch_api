-- +goose Up 
ALTER TABLE sessions
ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';

-- +goose Down
ALTER TABLE sessions
DROP COLUMN status;

