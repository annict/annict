-- name: GetEpisodeRecordByID :one
SELECT id, user_id, episode_id, work_id, record_id, body, rating, rating_state, created_at, updated_at
FROM episode_records
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: CountEpisodeRecordsByUserID :one
SELECT COUNT(*)
FROM episode_records
WHERE user_id = $1 AND deleted_at IS NULL;
