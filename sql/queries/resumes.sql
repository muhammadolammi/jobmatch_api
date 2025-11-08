-- name: CreateResume :one
INSERT INTO resumes (session_id, object_key, original_filename, mime, size_bytes, storage_provider, upload_status, storage_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)

RETURNING *;


-- name: GetResumes :one 
SELECT * FROM resumes;


-- name: ClearResumes :exec
DELETE  FROM resumes;

-- name: DeleteResumesBySession :exec
DELETE  FROM resumes WHERE session_id=$1;


-- name: GetResumesBySession :many 
SELECT * FROM resumes WHERE session_id=$1;

