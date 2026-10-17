-- name: DeleteWorkRecordsByUserID :one
-- ユーザーの作品の記録を最大batch_size件削除し、削除した件数を返す。
-- 記録に付いたいいねと、そのいいねで記録の持ち主 (削除するユーザー) に届いた通知も
-- 同じ文の中で削除する。いいね先の種類には、Rails版の別名 (AnimeRecord) で作られた
-- いいねも含める。論理削除されていない記録の分だけ、作品の記録数と本文のある記録数を
-- 減算する (Railsの再集計タスクと同じ定義)。updated_atは変えない。
-- 記録を参照するアクティビティとエピソードの記録 (review_id) は先に削除しておく
-- (外部キーがあるため)。
WITH deleted AS (
    DELETE FROM work_records
    WHERE id IN (
        SELECT target.id
        FROM work_records AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING work_records.id, work_records.work_id, work_records.body, work_records.deleted_at
),
deleted_likes AS (
    DELETE FROM likes
    WHERE likes.recipient_type IN ('WorkRecord', 'AnimeRecord')
        AND likes.recipient_id IN (SELECT deleted.id FROM deleted)
    RETURNING likes.id
),
deleted_notifications AS (
    DELETE FROM notifications
    WHERE notifications.user_id = sqlc.arg('user_id')
        AND notifications.trackable_type = 'Like'
        AND notifications.trackable_id IN (SELECT deleted_likes.id FROM deleted_likes)
    RETURNING notifications.id
),
counts AS (
    SELECT
        deleted.work_id,
        COUNT(*) AS records_count,
        COUNT(*) FILTER (WHERE deleted.body <> '') AS with_body_count
    FROM deleted
    WHERE deleted.deleted_at IS NULL
    GROUP BY deleted.work_id
),
decremented AS (
    UPDATE works
    SET
        work_records_count = works.work_records_count - counts.records_count,
        work_records_with_body_count = works.work_records_with_body_count - counts.with_body_count
    FROM counts
    WHERE works.id = counts.work_id
    RETURNING works.id
)
SELECT COUNT(*) FROM deleted;

-- name: DetachDeletedOAuthApplicationsFromWorkRecords :one
-- ユーザーが持つ論理削除済みのアプリケーションから投稿された他人の作品の記録を最大batch_size件、
-- アプリケーションとの紐づきを外し、外した件数を返す。投稿元のアプリケーションは表示に
-- 使われないため、作品の記録そのものと更新日時は変えない。
WITH detached AS (
    UPDATE work_records
    SET oauth_application_id = NULL
    WHERE id IN (
        SELECT target.id
        FROM work_records AS target
        WHERE target.oauth_application_id IN (
            SELECT oauth_applications.id
            FROM oauth_applications
            WHERE oauth_applications.owner_type = 'User'
                AND oauth_applications.owner_id = sqlc.arg('user_id')::bigint
                AND oauth_applications.deleted_at IS NOT NULL
        )
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING work_records.id
)
SELECT COUNT(*) FROM detached;
