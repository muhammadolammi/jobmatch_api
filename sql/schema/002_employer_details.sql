-- +goose Up 
CREATE TABLE employer_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_name TEXT UNIQUE NOT NULL , 
    company_website TEXT UNIQUE NOT NULL , 
    company_size INT NOT NULL,
    company_industry TEXT  NOT NULL,

    user_id UUID UNIQUE NOT NULL,
   constraint fk_employer_profiles_users
    foreign key (user_id) 
    REFERENCES users(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE employer_profiles;