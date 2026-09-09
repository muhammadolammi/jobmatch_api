-- name: CreateJobPost :one
INSERT INTO job_posts (
 employer_id, title, description )
VALUES ( $1, $2, $3)
RETURNING *;

-- name: GetEmployerJobPosts :many
SELECT * FROM job_posts 
WHERE employer_id = $1
ORDER BY created_at DESC;

-- name: GetJobPosts :many
SELECT * FROM job_posts 
ORDER BY created_at DESC;

-- name: GetJobPost :one 
SELECT * FROM job_posts 
WHERE id = $1;