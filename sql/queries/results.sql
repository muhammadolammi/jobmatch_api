
-- name: DeleteResultBySession :exec
DELETE  FROM results WHERE session_id=$1;
-- name: GetResultBySession :one 
SELECT * FROM results WHERE session_id=$1;
-- name: GetAllResults :one 
SELECT * FROM results ;


-- name: CreateResult :exec
INSERT INTO results (
result, session_id)
VALUES ( $1, $2);
