-- name: CreateUser :one
INSERT INTO users (name, role, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING id, name, role, email;

-- name: DeleteUserByName :exec
DELETE FROM users
WHERE name = $1;


-- name: ListUsers :many
SELECT id, name, role, email
FROM users
ORDER BY name;
