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
