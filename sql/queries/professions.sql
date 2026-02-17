-- name: GetProfessions :many
SELECT * FROM professions;


-- name: CreateProfession :one
INSERT INTO professions (
name  )
VALUES ( $1 )
RETURNING *;


-- name: ProfessionExists :one
SELECT EXISTS (
    SELECT 1
    FROM professions
    WHERE id = $1
);



-- name: GetUserProfessions :many
SELECT * FROM user_professions
WHERE user_id = $1;


-- name: CreateUserProfession :one
INSERT INTO user_professions (
user_id, profession_id  )
VALUES ( $1, $2 )
RETURNING *;

-- name: DeleteUserProfession :exec
DELETE FROM user_professions
WHERE user_id = $1 AND profession_id = $2;