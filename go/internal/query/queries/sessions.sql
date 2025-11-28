-- name: GetSessionByID :one
SELECT id, session_id, data, created_at, updated_at
FROM sessions
WHERE session_id = $1
LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (session_id, data, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
RETURNING id, session_id, data, created_at, updated_at;

-- name: UpdateSession :exec
UPDATE sessions
SET data = $2, updated_at = NOW()
WHERE session_id = $1;

-- name: TouchSession :exec
UPDATE sessions
SET updated_at = CLOCK_TIMESTAMP()
WHERE session_id = $1;