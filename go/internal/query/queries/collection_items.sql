-- name: DeleteCollectionItemsByUserID :one
-- ユーザーのコレクション項目と、ユーザーのコレクションに入っている項目を最大batch_size件削除し、
-- 削除した件数を返す。項目に付いたリアクションは外部キーのON DELETE CASCADEで消える。
WITH deleted AS (
    DELETE FROM collection_items
    WHERE id IN (
        SELECT target.id
        FROM collection_items AS target
        WHERE target.user_id = sqlc.arg('user_id')
            OR target.collection_id IN (
                SELECT collections.id
                FROM collections
                WHERE collections.user_id = sqlc.arg('user_id')
            )
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING collection_items.id
)
SELECT COUNT(*) FROM deleted;
