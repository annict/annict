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

-- name: AggregateDailyRecordCountsByUserID :many
-- Aggregates the number of records per day for the given user, after
-- converting watched_at from UTC to the supplied time zone. The result
-- only contains days that have at least one record; callers are expected
-- to zero-fill missing days when building a contiguous range.
--
-- watched_at is stored as a UTC timestamp (Rails convention) even though
-- the column type is `timestamp without time zone`, so it is interpreted
-- as UTC before being converted to the caller's time zone.
--
-- [Ja] watched_at を UTC から指定タイムゾーンへ変換した上で、ユーザー単位の
-- 日次レコード数を集計する。結果には記録のある日のみが含まれるため、連続した
-- 日付範囲を作るときは呼び出し元で 0 埋めする前提とする。
--
-- watched_at は `timestamp without time zone` 型だが Rails 規約に従って
-- UTC で保存されているため、いったん UTC として解釈してから指定タイムゾーン
-- に変換する。
SELECT
    (DATE((watched_at AT TIME ZONE 'UTC') AT TIME ZONE sqlc.arg(time_zone)::text))::date AS day,
    COUNT(*) AS count
FROM records
WHERE user_id = sqlc.arg(user_id)
  AND watched_at >= sqlc.arg(date_from)
  AND deleted_at IS NULL
GROUP BY day
ORDER BY day;
