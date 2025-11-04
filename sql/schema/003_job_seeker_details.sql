-- +goose Up 
CREATE TABLE job_seeker_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name TEXT UNIQUE NOT NULL , 
    last_name TEXT UNIQUE NOT NULL , 
    resume_url TEXT UNIQUE ,
    user_id UUID UNIQUE NOT NULL,
    constraint fk_job_seeker_profiles_users
    foreign key (user_id) 
    REFERENCES users(id)
);

-- +goose Down
DROP TABLE job_seeker_profiles;