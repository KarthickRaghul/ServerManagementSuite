-- name: CreateServerDevice :one
INSERT INTO server_devices (ip, tag, os, access_token)
VALUES ($1, $2, $3, $4)
RETURNING id, ip, tag, os, created_at;

-- name: DeleteServerDevice :exec
DELETE FROM server_devices 
WHERE ip = $1;

-- name: GetAllServerDevices :many
SELECT id, ip, tag, os, created_at , access_token
FROM server_devices 
ORDER BY created_at ASC;

-- name: GetServerDeviceByIP :one
SELECT id, ip, tag, os, access_token
FROM server_devices 
WHERE ip = $1;

