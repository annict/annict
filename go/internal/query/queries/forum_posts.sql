-- name: DeleteForumPostsByUserID :one
-- ユーザーが立てたスレッドを最大batch_size件削除し、削除した件数を返す。
-- スレッドのコメントと参加者は先に削除しておく (外部キーがあるため)。
WITH deleted AS (
    DELETE FROM forum_posts
    WHERE id IN (
        SELECT target.id
        FROM forum_posts AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING forum_posts.id
)
SELECT COUNT(*) FROM deleted;
