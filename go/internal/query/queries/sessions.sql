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

-- name: DeleteExpiredSessions :one
-- updated_atがlower_bound以上かつcutoffより古いセッションを最大batch_size件削除し、
-- 削除した件数と、削除した行のupdated_atの最大値を返す。PostgreSQLのDELETEはLIMITを
-- 取れないため、対象はupdated_atで並べたサブクエリで選び、
-- index_sessions_on_updated_atから古い順に読む。呼び出し元は返した最大値を次のバッチの
-- lower_boundに渡すことで、削除済みでVACUUMを待つインデックスエントリを先頭から読み
-- 直さずに済む。lower_boundを「以上」にしているのは、同じupdated_atの行がバッチの境目で
-- 分かれても取りこぼさないため。SKIP LOCKEDにより、並行実行時は他方がロック中の行を
-- 飛ばして次へ進める。付けない場合、後発は待たされた末に0件を削除することになり、滞留が
-- 残っていてもそこで消化が止まる。削除が0件のときの最大値はlower_boundを返す。
WITH deleted AS (
    DELETE FROM sessions
    WHERE id IN (
        SELECT expired.id
        FROM sessions AS expired
        WHERE expired.updated_at >= sqlc.arg('lower_bound')
            AND expired.updated_at < sqlc.arg('cutoff')
        ORDER BY expired.updated_at
        LIMIT sqlc.arg('batch_size')
        FOR UPDATE SKIP LOCKED
    )
    RETURNING sessions.updated_at
)
SELECT
    COUNT(*) AS deleted_count,
    COALESCE(MAX(deleted.updated_at), sqlc.arg('lower_bound'))::timestamp AS max_updated_at
FROM deleted;
