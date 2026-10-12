-- name: DeleteStatusesByUserID :one
-- ユーザーのステータスを最大batch_size件削除し、削除した件数を返す。
-- ステータスに付いたいいねと、そのいいねでステータスの持ち主 (削除するユーザー) に
-- 届いた通知も同じ文の中で削除する。持ち主宛ての通知に絞ることで、notificationsは
-- user_idのインデックスから引ける。ステータスを参照するアクティビティとライブラリ
-- エントリは先に削除しておく (外部キーがあるため)。
WITH deleted AS (
    DELETE FROM statuses
    WHERE id IN (
        SELECT target.id
        FROM statuses AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING statuses.id
),
deleted_likes AS (
    DELETE FROM likes
    WHERE likes.recipient_type = 'Status'
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

-- name: DetachDeletedOAuthApplicationsFromStatuses :one
-- ユーザーが持つ論理削除済みのアプリケーションから投稿された他人のステータスを最大batch_size件、
-- アプリケーションとの紐づきを外し、外した件数を返す。投稿元のアプリケーションは表示に
-- 使われないため、ステータスそのものと更新日時は変えない。
WITH detached AS (
    UPDATE statuses
    SET oauth_application_id = NULL
    WHERE id IN (
        SELECT target.id
        FROM statuses AS target
        WHERE target.oauth_application_id IN (
            SELECT oauth_applications.id
            FROM oauth_applications
            WHERE oauth_applications.owner_type = 'User'
                AND oauth_applications.owner_id = sqlc.arg('user_id')::bigint
                AND oauth_applications.deleted_at IS NOT NULL
        )
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING statuses.id
)
SELECT COUNT(*) FROM detached;
