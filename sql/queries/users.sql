-- name: GetUsers :many
SELECT * FROM users;


-- name: CreateUser :one
INSERT INTO users (
email, role,password  )
VALUES ( $1, $2, $3 )
RETURNING *;
-- name: DeleteUser :exec
DELETE FROM users
WHERE id=$1;

-- name: UserExists :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE email = $1
);


-- name: GetUserWithEmail :one
SELECT * FROM users WHERE $1=email;
-- name: GetUser :one
SELECT * FROM users WHERE $1=id;


-- name: UpdatePassword :exec
UPDATE users
SET 
  password = $1
WHERE email = $2;




-- name: CreateEmployer :one
INSERT INTO employer_profiles (
user_id, company_name,company_website,company_size,company_industry  )
VALUES ( $1, $2, $3,$4,$5 )
RETURNING *;

-- name: CreateJobSeeker :one
INSERT INTO job_seeker_profiles (
user_id, first_name,last_name  )
VALUES ( $1, $2, $3)
RETURNING *;
