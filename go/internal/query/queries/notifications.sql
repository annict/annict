-- name: DeleteNotificationsByActionUserID :one
-- ユーザーが起こした通知 (ユーザーのいいね・コメント・フォローで他人に届いた通知) を
-- 最大batch_size件削除し、削除した件数を返す。他人宛ての通知のうち未読の分だけ、
-- 宛先のusers.notifications_count (未読数) を減算する。updated_atは変えない。
-- 通知はaction_user_idのインデックスから引く。通知のきっかけになったいいね・コメントより
-- 先に削除し、他人の通知一覧に参照先の無い通知を残さないようにする。
WITH deleted AS (
    DELETE FROM notifications
    WHERE id IN (
        SELECT target.id
        FROM notifications AS target
        WHERE target.action_user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING notifications.user_id, notifications.read
),
unread_counts AS (
    SELECT deleted.user_id, COUNT(*) AS unread_count
    FROM deleted
    WHERE NOT deleted.read
        AND deleted.user_id <> sqlc.arg('user_id')
    GROUP BY deleted.user_id
),
decremented AS (
    UPDATE users
    SET notifications_count = users.notifications_count - unread_counts.unread_count
    FROM unread_counts
    WHERE users.id = unread_counts.user_id
    RETURNING users.id
)
SELECT COUNT(*) FROM deleted;

-- name: DeleteNotificationsByUserID :one
-- ユーザー宛ての通知を最大batch_size件削除し、削除した件数を返す。
-- 未読数はユーザー自身の行が持ち、ユーザー行ごと削除するため減算しない。
WITH deleted AS (
    DELETE FROM notifications
    WHERE id IN (
        SELECT target.id
        FROM notifications AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING notifications.id
)
SELECT COUNT(*) FROM deleted;
