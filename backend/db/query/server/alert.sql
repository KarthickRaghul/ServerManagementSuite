-- name: CreateAlert :one
INSERT INTO alerts (host, severity, content, status)
VALUES ($1, $2, $3, 'notseen')
RETURNING id, host, severity, content, status, time;

-- name: GetAlertsByHost :many
SELECT id, host, severity, content, status, time
FROM alerts 
WHERE host = $1
ORDER BY time DESC;

-- name: GetAllAlerts :many
SELECT id, host, severity, content, status, time
FROM alerts 
ORDER BY time DESC
LIMIT $1;

-- name: GetUnseenAlerts :many
SELECT id, host, severity, content, status, time
FROM alerts 
WHERE status = 'notseen'
ORDER BY time DESC
LIMIT $1;

-- name: GetUnseenAlertsByHost :many
SELECT id, host, severity, content, status, time
FROM alerts 
WHERE host = $1 AND status = 'notseen'
ORDER BY time DESC;

-- name: MarkAlertAsSeen :exec
UPDATE alerts 
SET status = 'seen'
WHERE id = $1;

-- name: MarkMultipleAlertsAsSeen :exec
UPDATE alerts 
SET status = 'seen'
WHERE id = ANY($1::int[]);

-- name: DeleteAlertById :exec
DELETE FROM alerts 
WHERE id = $1;

-- name: DeleteMultipleAlerts :exec
DELETE FROM alerts 
WHERE id = ANY($1::int[]);

-- name: GetRecentAlerts :many
SELECT id, host, severity, content, status, time
FROM alerts 
WHERE time > $1
ORDER BY time DESC;

-- name: DeleteOldAlerts :exec
delete from alerts 
where time < $1;

