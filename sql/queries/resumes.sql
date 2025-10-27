-- name: CreateResume :one
INSERT INTO resumes (
file_name, text, session_id)
VALUES ( $1, $2, $3)
RETURNING *;


-- name: GetResumes :one 
SELECT * FROM resumes;


-- name: ClearResumes :exec
DELETE  FROM resumes;

-- name: DeleteResumesBySession :exec
DELETE  FROM resumes WHERE session_id=$1;