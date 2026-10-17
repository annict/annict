-- name: GetActivityGroupByID :one
SELECT id, user_id, itemable_type, single, activities_count, created_at, updated_at
FROM activity_groups
WHERE id = $1
LIMIT 1;

-- name: GetActivityGroupByUserAndType :one
SELECT id, user_id, itemable_type, single, activities_count, created_at, updated_at
FROM activity_groups
WHERE user_id = $1 AND itemable_type = $2 AND single = $3
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteActivityGroupsByUserID :one
-- ユーザーのアクティビティグループを最大batch_size件削除し、削除した件数を返す。
-- グループに属するアクティビティは先に削除しておく (activitiesからの外部キーがあるため)。
WITH deleted AS (
    DELETE FROM activity_groups
    WHERE id IN (
        SELECT target.id
        FROM activity_groups AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING activity_groups.id
)
SELECT COUNT(*) FROM deleted;
