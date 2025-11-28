-- name: GetRecordByID :one
SELECT id, user_id, work_id, aasm_state, impressions_count, created_at, updated_at, watched_at
FROM records
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetRecordByUserAndWork :one
SELECT id, user_id, work_id, aasm_state, impressions_count, created_at, updated_at, watched_at
FROM records
WHERE user_id = $1 AND work_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: CountRecordsByUserID :one
SELECT COUNT(*)
FROM records
WHERE user_id = $1 AND deleted_at IS NULL;
