-- +goose Up
ALTER TABLE sessions
ADD COLUMN job_title TEXT NOT NULL DEFAULT 'empty',
ADD COLUMN job_description TEXT NOT NULL DEFAULT 'empty';

-- +goose Down
ALTER TABLE sessions
DROP COLUMN job_description,
DROP COLUMN job_title;
