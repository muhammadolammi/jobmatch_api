


-- name: CreateOrUpdateAnalysesResults :exec
INSERT INTO analyses_results (
results, session_id)
VALUES ( $1, $2)
ON CONFLICT (session_id)
DO UPDATE SET
    results = EXCLUDED.results,
    updated_at = CURRENT_TIMESTAMP
;

-- name: GetAnalysesResultsBySession :one 
SELECT * FROM analyses_results WHERE session_id=$1;