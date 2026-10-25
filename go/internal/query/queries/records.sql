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
-- watched_atをUTCから指定タイムゾーンへ変換した上で、ユーザー単位の
-- 日次レコード数を集計する。結果には記録のある日のみが含まれるため、連続した
-- 日付範囲を作るときは呼び出し元で0埋めする前提とする。
--
-- watched_atは `timestamp without time zone` 型だがRails規約に従って
-- UTCで保存されているため、いったんUTCとして解釈してから指定タイムゾーン
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

-- name: DeleteRecordsByUserID :one
-- ユーザーの記録を最大batch_size件削除し、削除した件数を返す。
-- 論理削除されていない記録の分だけ、同じ文の中でworks.records_countを減算する
-- (Railsの再集計タスクと同じ定義)。updated_atは変えない。記録を参照するエピソードの
-- 記録と作品の記録は先に削除しておく (外部キーがあるため)。
WITH deleted AS (
    DELETE FROM records
    WHERE id IN (
        SELECT target.id
        FROM records AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING records.work_id, records.deleted_at
),
counts AS (
    SELECT deleted.work_id, COUNT(*) AS records_count
    FROM deleted
    WHERE deleted.deleted_at IS NULL
    GROUP BY deleted.work_id
),
decremented AS (
    UPDATE works
    SET records_count = works.records_count - counts.records_count
    FROM counts
    WHERE works.id = counts.work_id
    RETURNING works.id
)
SELECT COUNT(*) FROM deleted;
