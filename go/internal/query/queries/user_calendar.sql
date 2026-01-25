-- name: GetUserCalendarInfo :one
-- ユーザーのカレンダー情報を取得します
SELECT
    u.username,
    u.time_zone,
    u.locale
FROM users u
WHERE LOWER(u.username) = LOWER(@username)
    AND u.deleted_at IS NULL
LIMIT 1;

-- name: GetLibraryEntryProgramIDs :many
-- ユーザーの視聴リスト（見たい・見てる）からprogram_idを取得します
SELECT
    le.program_id,
    le.watched_episode_ids
FROM library_entries le
JOIN statuses s ON s.id = le.status_id
WHERE le.user_id = @user_id
    AND s.kind IN (1, 2)
    AND le.program_id IS NOT NULL;

-- name: GetCalendarSlots :many
-- カレンダー用の放送枠を取得します
-- 現在時刻から7日後までの未視聴エピソードを対象とします
SELECT
    s.id,
    s.started_at,
    s.work_id,
    w.title AS work_title,
    w.title_en AS work_title_en,
    s.episode_id,
    e.title AS episode_title,
    COALESCE(e.number, '') AS episode_number,
    c.name AS channel_name
FROM slots s
JOIN works w ON w.id = s.work_id
JOIN episodes e ON e.id = s.episode_id
JOIN channels c ON c.id = s.channel_id
WHERE s.deleted_at IS NULL
    AND w.deleted_at IS NULL
    AND e.deleted_at IS NULL
    AND s.program_id = ANY(@program_ids::bigint[])
    AND s.started_at >= @started_at_from
    AND s.started_at <= @started_at_to
    AND s.episode_id IS NOT NULL
ORDER BY s.started_at ASC;

-- name: GetCalendarWorks :many
-- カレンダー用の作品（放送開始日）を取得します
SELECT
    w.id,
    w.title,
    w.title_en,
    w.started_on
FROM works w
JOIN library_entries le ON le.work_id = w.id
JOIN statuses s ON s.id = le.status_id
WHERE le.user_id = @user_id
    AND s.kind IN (1, 2)
    AND w.deleted_at IS NULL
    AND w.started_on IS NOT NULL;
