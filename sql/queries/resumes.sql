-- name: CreateResume :one
INSERT INTO resumes (
file_name, text, session_id)
VALUES ( $1, $2, $3)

ON CONFLICT (session_id)
DO UPDATE SET text = EXCLUDED.text
RETURNING *;


-- name: GetResumes :one 
SELECT * FROM resumes;


-- name: ClearResumes :exec
DELETE  FROM resumes;

-- name: DeleteResumesBySession :exec
DELETE  FROM resumes WHERE session_id=$1;