-- name: DeleteDbCommentsByUserID :one
-- ユーザーが書いたAnnict DBのコメントを最大batch_size件削除し、削除した件数を返す。
WITH deleted AS (
    DELETE FROM db_comments
    WHERE id IN (
        SELECT target.id
        FROM db_comments AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING db_comments.id
)
SELECT COUNT(*) FROM deleted;
