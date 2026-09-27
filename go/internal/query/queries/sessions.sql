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

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE session_id = $1;

-- name: DeleteExpiredSessions :execrows
-- updated_atがcutoffより古いセッションを最大batch_size件削除する。PostgreSQLの
-- DELETEはLIMITを取れないため、対象はupdated_atで並べたサブクエリで選び、
-- index_sessions_on_updated_atから古い順に読む。SKIP LOCKEDにより、並行実行時は他方が
-- ロック中の行を飛ばして次へ進める。付けない場合、後発は待たされた末に0件を削除すること
-- になり、滞留が残っていてもそこで消化が止まる。
DELETE FROM sessions
WHERE id IN (
    SELECT expired.id
    FROM sessions AS expired
    WHERE expired.updated_at < sqlc.arg('cutoff')
    ORDER BY expired.updated_at
    LIMIT sqlc.arg('batch_size')
    FOR UPDATE SKIP LOCKED
);
