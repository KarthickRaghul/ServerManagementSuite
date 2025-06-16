-- name: GetUserByName :one
SELECT id, name, role, email, password_hash 
FROM users 
WHERE name = $1;

-- name: SaveRefreshToken :exec
INSERT INTO user_sessions (username, refresh_token, expires_at)
VALUES ($1, $2, $3);

-- name: GetRefreshToken :one
SELECT username, expires_at 
FROM user_sessions 
WHERE refresh_token = $1 AND expires_at > now();

-- name: DeleteRefreshToken :exec
DELETE FROM user_sessions 
WHERE refresh_token = $1;

-- name: CleanExpiredSessions :exec
DELETE FROM user_sessions 
WHERE expires_at < now();

-- name: GetValidRefreshTokenByUser :one
SELECT refresh_token, expires_at, created_at
FROM user_sessions 
WHERE username = $1 AND expires_at > now()
ORDER BY created_at DESC
LIMIT 1;


