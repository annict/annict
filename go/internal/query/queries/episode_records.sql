-- name: GetEpisodeRecordByID :one
SELECT id, user_id, episode_id, work_id, record_id, body, rating, rating_state, created_at, updated_at
FROM episode_records
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: CountEpisodeRecordsByUserID :one
SELECT COUNT(*)
FROM episode_records
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: DeleteEpisodeRecordsByUserID :one
-- ユーザーのエピソードの記録を最大batch_size件削除し、削除した件数を返す。
-- 記録に付いたいいねと、そのいいねで記録の持ち主 (削除するユーザー) に届いた通知も
-- 同じ文の中で削除する。論理削除されていない記録の分だけ、エピソードの記録数と
-- 本文のある記録数を減算する (Railsの再集計タスクと同じ定義)。updated_atは変えない。
-- 記録を参照するアクティビティは先に削除しておく (外部キーがあるため)。記録に付いた
-- コメントは外部キーのON DELETE CASCADEでも消えるが、コメントのいいね・通知を
-- 残さないよう先の段階で削除しておく。
WITH deleted AS (
    DELETE FROM episode_records
    WHERE id IN (
        SELECT target.id
        FROM episode_records AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING episode_records.id, episode_records.episode_id, episode_records.body, episode_records.deleted_at
),
deleted_likes AS (
    DELETE FROM likes
    WHERE likes.recipient_type = 'EpisodeRecord'
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
        deleted.episode_id,
        COUNT(*) AS records_count,
        COUNT(*) FILTER (WHERE deleted.body IS NOT NULL AND deleted.body <> '') AS bodies_count
    FROM deleted
    WHERE deleted.deleted_at IS NULL
    GROUP BY deleted.episode_id
),
decremented AS (
    UPDATE episodes
    SET
        episode_records_count = episodes.episode_records_count - counts.records_count,
        episode_record_bodies_count = episodes.episode_record_bodies_count - counts.bodies_count
    FROM counts
    WHERE episodes.id = counts.episode_id
    RETURNING episodes.id
)
SELECT COUNT(*) FROM deleted;

-- name: DetachDeletedOAuthApplicationsFromEpisodeRecords :one
-- ユーザーが持つ論理削除済みのアプリケーションから投稿された他人のエピソードの記録を最大batch_size件、
-- アプリケーションとの紐づきを外し、外した件数を返す。投稿元のアプリケーションは表示に
-- 使われないため、エピソードの記録そのものと更新日時は変えない。
WITH detached AS (
    UPDATE episode_records
    SET oauth_application_id = NULL
    WHERE id IN (
        SELECT target.id
        FROM episode_records AS target
        WHERE target.oauth_application_id IN (
            SELECT oauth_applications.id
            FROM oauth_applications
            WHERE oauth_applications.owner_type = 'User'
                AND oauth_applications.owner_id = sqlc.arg('user_id')::bigint
                AND oauth_applications.deleted_at IS NOT NULL
        )
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING episode_records.id
)
SELECT COUNT(*) FROM detached;
