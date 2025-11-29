-- name: GetActivityByID :one
SELECT id, user_id, trackable_id, trackable_type, episode_record_id, work_id, episode_id, activity_group_id, created_at, updated_at
FROM activities
WHERE id = $1
LIMIT 1;

-- name: CountActivitiesByUserID :one
SELECT COUNT(*)
FROM activities
WHERE user_id = $1;
