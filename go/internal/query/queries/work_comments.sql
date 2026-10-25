-- name: DeleteWorkCommentsByUserID :one
-- ユーザーが書いた作品へのコメントを最大batch_size件削除し、削除した件数を返す。
WITH deleted AS (
    DELETE FROM work_comments
    WHERE id IN (
        SELECT target.id
        FROM work_comments AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING work_comments.id
)
SELECT COUNT(*) FROM deleted;
