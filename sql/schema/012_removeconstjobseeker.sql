-- +goose Up
ALTER TABLE job_seeker_profiles DROP CONSTRAINT IF EXISTS job_seeker_profiles_first_name_key;
ALTER TABLE job_seeker_profiles DROP CONSTRAINT IF EXISTS job_seeker_profiles_last_name_key;

-- +goose Down
ALTER TABLE job_seeker_profiles ADD CONSTRAINT job_seeker_profiles_first_name_key UNIQUE (first_name);
ALTER TABLE job_seeker_profiles ADD CONSTRAINT job_seeker_profiles_last_name_key UNIQUE (last_name);