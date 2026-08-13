-- name: ListDBEpisodes :many
-- The preceding episode is derived from the neighbouring row in ascending sort_number
-- order rather than read from episodes.prev_episode_id. The window runs over the work's
-- whole list inside the CTE, before LIMIT / OFFSET narrow it to one page, so the last row
-- of a page still names the episode that lands on the next page.
--
-- [Ja] 直前のエピソードは episodes.prev_episode_id を読まず、sort_number 昇順の隣接行から
-- 導出する。ウィンドウは CTE の中で作品の一覧全体に対して評価され、LIMIT / OFFSET が
-- 1 ページに絞り込む前に確定するため、ページ末尾の行も次ページに載るエピソードを指せる。
WITH work_episodes AS (
    SELECT
        e.id,
        e.work_id,
        e.number,
        e.raw_number,
        e.sort_number,
        e.title,
        e.title_ro,
        e.title_en,
        e.episode_records_count,
        e.unpublished_at,
        e.deleted_at,
        LAG(e.id) OVER (ORDER BY e.sort_number, e.id) AS prev_episode_id
    FROM episodes e
    WHERE e.work_id = sqlc.arg('work_id')
        AND e.deleted_at IS NULL
)
SELECT
    we.id,
    we.work_id,
    we.number,
    we.raw_number,
    we.sort_number,
    we.title,
    we.title_ro,
    we.title_en,
    we.episode_records_count,
    we.unpublished_at,
    we.deleted_at,
    prev.number AS prev_number,
    prev.raw_number AS prev_raw_number
FROM work_episodes we
LEFT JOIN episodes prev ON prev.id = we.prev_episode_id
ORDER BY we.sort_number DESC, we.id DESC
LIMIT sqlc.arg('per_page')
OFFSET sqlc.arg('page_offset')::bigint;

-- name: CountDBEpisodes :one
SELECT COUNT(*)
FROM episodes e
WHERE e.work_id = sqlc.arg('work_id')
    AND e.deleted_at IS NULL;

-- name: GetEpisodeForEditByID :one
-- The edit form reads the episode's editable columns together with the two the page needs
-- from its parent work: the title for the heading and no_episodes for the shared work
-- subnav. One row covers both, so opening the form costs a single round trip.
--
-- Deleted episodes are excluded by deleted_at (the Rails Episode.without_deleted.find the
-- edit action uses), and works are filtered the same way as on the episode list, so an
-- episode whose work is gone is not editable through a page whose heading and subnav point
-- at that work.
--
-- updated_at is the version the form carries in a hidden field so the update can reject a
-- submit made against a stale read.
--
-- [Ja] 編集フォームは、エピソードの編集対象カラムと、ページが親作品から必要とする 2 つの
-- カラム (見出しに使う title と、共有の作品サブナビが使う no_episodes) を一緒に読む。
-- 1 行で両方を賄うため、フォームを開くのに往復は 1 回で済む。
--
-- 削除済みエピソードは deleted_at で除外し (編集アクションが使う Rails の
-- Episode.without_deleted.find と同じ)、作品もエピソード一覧と同じ条件で絞る。作品が
-- 失われたエピソードを、その作品を見出しとサブナビで指すページから編集させないため。
--
-- updated_at はフォームが hidden で持ち回る版。古い読み取りに対する送信を更新側で
-- 却下できるようにする。
SELECT
    e.id,
    e.work_id,
    e.number,
    e.raw_number,
    e.sort_number,
    e.title,
    e.title_en,
    e.updated_at,
    w.title AS work_title,
    w.no_episodes AS work_no_episodes
FROM episodes e
INNER JOIN works w ON w.id = e.work_id
WHERE e.id = $1
    AND e.deleted_at IS NULL
    AND w.deleted_at IS NULL;

-- name: ListEpisodesForAnimeSyncByIDs :many
SELECT
    e.id,
    e.work_id,
    e.title,
    e.title_ro,
    e.title_en,
    e.number,
    e.sort_number,
    e.raw_number,
    e.status,
    e.archive_message,
    e.unpublished_at,
    e.deleted_at,
    e.anime_id,
    w.anime_id AS parent_anime_id
FROM episodes e
JOIN works w ON e.work_id = w.id
WHERE e.id = ANY($1::bigint[])
ORDER BY e.id;

-- name: ListEpisodeIDsAfter :many
SELECT id
FROM episodes
WHERE id > sqlc.arg('after_id')
ORDER BY id
LIMIT sqlc.arg('batch_size');

-- name: UpdateEpisodeAnimeID :exec
UPDATE episodes
SET anime_id = $2
WHERE id = $1;

-- name: CreateEpisode :one
-- anime_id is written with the row rather than patched afterwards: the bulk create inserts
-- the episode's anime before the episode itself, so the mapping column is already known.
-- prev_episode_id names the episode with the greatest sort_number at insert time, the value
-- the Rails after_create callback assigns; the Annict DB list derives the preceding episode
-- from sort_number order instead, but the public episode page and the GraphQL API still read
-- the column. The data-modifying CTE also records the same episodes.create DB activity as
-- Rails save_and_create_activity!, using the inserted row as parameters.new.
--
-- [Ja] anime_id は後から書き戻さず行と一緒に書く。一括作成はエピソード本体より先にその
-- anime を挿入するため、マッピングカラムの値が既に分かっているため。prev_episode_id には
-- 挿入時点で sort_number が最大のエピソードを入れる (Rails の after_create コールバックが
-- 入れるのと同じ値)。Annict DB の一覧は直前のエピソードを sort_number 順から導出するが、
-- 公開側のエピソードページと GraphQL API は今もこのカラムを読む。データ変更 CTE はさらに、
-- Rails の save_and_create_activity! と同じ episodes.create の DB 活動を、挿入行を
-- parameters.new として記録する。
WITH created_episode AS (
    INSERT INTO episodes (
        work_id,
        number,
        raw_number,
        title,
        sort_number,
        prev_episode_id,
        anime_id,
        created_at,
        updated_at
    ) VALUES (
        sqlc.arg('work_id'),
        sqlc.narg('number'),
        sqlc.narg('raw_number'),
        sqlc.narg('title'),
        sqlc.arg('sort_number'),
        sqlc.narg('prev_episode_id'),
        sqlc.narg('anime_id'),
        NOW(),
        NOW()
    )
    RETURNING *
), created_activity AS (
    INSERT INTO db_activities (
        user_id,
        trackable_id,
        trackable_type,
        action,
        parameters,
        created_at,
        updated_at,
        root_resource_id,
        root_resource_type
    )
    SELECT
        sqlc.arg('user_id'),
        ce.id,
        'Episode',
        'episodes.create',
        json_build_object('new', row_to_json(ce)),
        NOW(),
        NOW(),
        ce.work_id,
        'Work'
    FROM created_episode ce
    RETURNING id
)
SELECT ce.id
FROM created_episode ce
CROSS JOIN created_activity ca;
