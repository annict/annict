-- name: DeleteWorkTaggablesByUserID :one
-- ユーザーが使ったタグの記録 (タグの説明を含む) を最大batch_size件削除し、削除した件数を返す。
-- タグ (work_tags) は他のユーザーと共有するため残す。
WITH deleted AS (
    DELETE FROM work_taggables
    WHERE id IN (
        SELECT target.id
        FROM work_taggables AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING work_taggables.id
)
SELECT COUNT(*) FROM deleted;
