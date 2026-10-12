-- name: DeleteCommentsOnEpisodeRecordsByUserID :one
-- ユーザーのエピソードの記録に付いたコメント (書いたのが誰かを問わない) を最大
-- batch_size件削除し、削除した件数を返す。同じ文の中で次も削除する。
-- - コメントでユーザーに届いた通知 (コメントの通知は記録の持ち主宛て)
-- - コメントに付いたいいねと、そのいいねでコメントを書いた人に届いた通知
-- 他人宛ての通知のうち未読の分だけ、宛先のusers.notifications_count (未読数) を減算する。
-- 通知は宛先 (user_id) で絞り、user_idのインデックスから引く。
WITH deleted AS (
    DELETE FROM comments
    WHERE id IN (
        SELECT target.id
        FROM comments AS target
        INNER JOIN episode_records ON episode_records.id = target.episode_record_id
        WHERE episode_records.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING comments.id, comments.user_id
),
deleted_likes AS (
    DELETE FROM likes
    WHERE likes.recipient_type = 'Comment'
        AND likes.recipient_id IN (SELECT deleted.id FROM deleted)
    RETURNING likes.id, likes.recipient_id
),
deleted_comment_notifications AS (
    DELETE FROM notifications
    WHERE notifications.user_id = sqlc.arg('user_id')
        AND notifications.trackable_type = 'Comment'
        AND notifications.trackable_id IN (SELECT deleted.id FROM deleted)
    RETURNING notifications.id
),
deleted_like_notifications AS (
    DELETE FROM notifications
    USING deleted_likes
    INNER JOIN deleted ON deleted.id = deleted_likes.recipient_id
    WHERE notifications.user_id = deleted.user_id
        AND notifications.trackable_type = 'Like'
        AND notifications.trackable_id = deleted_likes.id
    RETURNING notifications.user_id, notifications.read
),
unread_counts AS (
    SELECT deleted_like_notifications.user_id, COUNT(*) AS unread_count
    FROM deleted_like_notifications
    WHERE NOT deleted_like_notifications.read
        AND deleted_like_notifications.user_id <> sqlc.arg('user_id')
    GROUP BY deleted_like_notifications.user_id
),
decremented AS (
    UPDATE users
    SET notifications_count = users.notifications_count - unread_counts.unread_count
    FROM unread_counts
    WHERE users.id = unread_counts.user_id
    RETURNING users.id
)
SELECT COUNT(*) FROM deleted;

-- name: DeleteCommentsByUserID :one
-- ユーザーが書いたコメントを最大batch_size件削除し、削除した件数を返す。
-- コメントに付いたいいねも同じ文の中で削除し、コメント先のエピソードの記録の
-- comments_countを削除した件数だけ減算する。updated_atは変えない。
-- コメントで他人に届いた通知と、コメントへのいいねでユーザーに届いた通知は、
-- 先の段階 (ユーザーが起こした通知・ユーザー宛ての通知の削除) で削除しておく。
-- ユーザー自身の記録に書いたコメントは、記録の段階で削除済みになる。
WITH deleted AS (
    DELETE FROM comments
    WHERE id IN (
        SELECT target.id
        FROM comments AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING comments.id, comments.episode_record_id
),
deleted_likes AS (
    DELETE FROM likes
    WHERE likes.recipient_type = 'Comment'
        AND likes.recipient_id IN (SELECT deleted.id FROM deleted)
    RETURNING likes.id
),
counts AS (
    SELECT deleted.episode_record_id, COUNT(*) AS comments_count
    FROM deleted
    GROUP BY deleted.episode_record_id
),
decremented AS (
    UPDATE episode_records
    SET comments_count = episode_records.comments_count - counts.comments_count
    FROM counts
    WHERE episode_records.id = counts.episode_record_id
    RETURNING episode_records.id
)
SELECT COUNT(*) FROM deleted;
