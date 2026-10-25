-- name: DeleteMultipleEpisodeRecordsByUserID :one
-- ユーザーの複数エピソードの記録を最大batch_size件削除し、削除した件数を返す。
-- 記録に付いたいいねと、そのいいねで記録の持ち主 (削除するユーザー) に届いた通知も
-- 同じ文の中で削除する。記録を参照するアクティビティとエピソードの記録は先に削除して
-- おく (外部キーがあるため)。
WITH deleted AS (
    DELETE FROM multiple_episode_records
    WHERE id IN (
        SELECT target.id
        FROM multiple_episode_records AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING multiple_episode_records.id
),
deleted_likes AS (
    DELETE FROM likes
    WHERE likes.recipient_type = 'MultipleEpisodeRecord'
        AND likes.recipient_id IN (SELECT deleted.id FROM deleted)
    RETURNING likes.id
),
deleted_notifications AS (
    DELETE FROM notifications
    WHERE notifications.user_id = sqlc.arg('user_id')
        AND notifications.trackable_type = 'Like'
        AND notifications.trackable_id IN (SELECT deleted_likes.id FROM deleted_likes)
    RETURNING notifications.id
)
SELECT COUNT(*) FROM deleted;
