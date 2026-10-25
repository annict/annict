-- name: DeleteCollectionsByUserID :one
-- ユーザーのコレクションを最大batch_size件削除し、削除した件数を返す。
-- コレクションに付いた他人のいいねも同じ文の中で削除する。いいねでユーザーに届いた通知は
-- 通知の段階で削除済みのため扱わない。コレクション項目は先に削除しておく (外部キーがあるため)。
WITH deleted AS (
    DELETE FROM collections
    WHERE id IN (
        SELECT target.id
        FROM collections AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING collections.id
),
deleted_likes AS (
    DELETE FROM likes
    WHERE likes.recipient_type = 'Collection'
        AND likes.recipient_id IN (SELECT deleted.id FROM deleted)
    RETURNING likes.id
)
SELECT COUNT(*) FROM deleted;
