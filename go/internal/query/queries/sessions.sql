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
-- Deletes at most batch_size sessions whose updated_at is older than cutoff. PostgreSQL
-- does not accept LIMIT on DELETE, so the rows are picked by a subquery ordered by
-- updated_at, which reads the oldest ones off index_sessions_on_updated_at. SKIP LOCKED
-- lets a concurrent run step over the rows another run already holds: without it the
-- second run would wait, then delete zero rows and stop while a backlog remains.
--
-- [Ja] updated_at が cutoff より古いセッションを最大 batch_size 件削除する。PostgreSQL の
-- DELETE は LIMIT を取れないため、対象は updated_at で並べたサブクエリで選び、
-- index_sessions_on_updated_at から古い順に読む。SKIP LOCKED により、並行実行時は他方が
-- ロック中の行を飛ばして次へ進める。付けない場合、後発は待たされた末に 0 件を削除すること
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
