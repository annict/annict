-- name: GetActivityByID :one
SELECT id, user_id, trackable_id, trackable_type, episode_record_id, work_id, episode_id, activity_group_id, created_at, updated_at
FROM activities
WHERE id = $1
LIMIT 1;

-- name: CountActivitiesByUserID :one
SELECT COUNT(*)
FROM activities
WHERE user_id = $1;

-- name: DeleteActivitiesByUserID :one
-- ユーザーのアクティビティを最大batch_size件削除し、削除した件数を返す。
WITH deleted AS (
    DELETE FROM activities
    WHERE id IN (
        SELECT target.id
        FROM activities AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING activities.id
)
SELECT COUNT(*) FROM deleted;
