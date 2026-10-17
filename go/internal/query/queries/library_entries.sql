-- name: DeleteLibraryEntriesByUserID :one
-- ユーザーのライブラリエントリを最大batch_size件削除し、削除した件数を返す。
-- 見たい・見てる・見た (statuses.kindが1・2・3) のエントリの分だけ、同じ文の中で
-- works.watchers_countを減算する。1文にまとめることで、削除と減算が必ず一緒に
-- 確定し、再実行しても二重に減算しない。ステータスは後から削除するため、この時点では
-- 参照できる。updated_atはRailsのカウンタの更新と同じく変えない。
WITH deleted AS (
    DELETE FROM library_entries
    WHERE id IN (
        SELECT target.id
        FROM library_entries AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING library_entries.work_id, library_entries.status_id
),
watchers AS (
    SELECT deleted.work_id, COUNT(*) AS watchers_count
    FROM deleted
    INNER JOIN statuses ON statuses.id = deleted.status_id
    WHERE statuses.kind IN (1, 2, 3)
    GROUP BY deleted.work_id
),
decremented AS (
    UPDATE works
    SET watchers_count = works.watchers_count - watchers.watchers_count
    FROM watchers
    WHERE works.id = watchers.work_id
    RETURNING works.id
)
SELECT COUNT(*) FROM deleted;
