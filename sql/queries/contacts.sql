

-- name: GetContactDepartments :many
SELECT * FROM contact_departments;


-- name: CreateContactMessage :one
INSERT INTO contact_messages (
first_name, last_name,email,contact_department_id, message )
VALUES ( $1, $2, $3, $4, $5 )
RETURNING *; 

-- name: GetEmailLastHourContactMessages :one
SELECT COUNT(*)
FROM contact_messages
WHERE email = $1
AND created_at >= NOW() - INTERVAL '1 hour'; 

-- name: GetEmailLast24HourContactMessages :one
SELECT COUNT(*)
FROM contact_messages
WHERE email = $1
AND created_at >= NOW() - INTERVAL '24 hours';