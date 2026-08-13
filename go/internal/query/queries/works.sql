-- name: GetPopularWorks :many
SELECT
    w.id,
    w.title,
    w.title_en,
    w.recommended_image_url,
    wi.image_data,
    w.watchers_count,
    w.season_year,
    w.season_name,
    w.created_at
FROM works w
LEFT JOIN work_images wi ON w.id = wi.work_id
WHERE w.watchers_count > 0
ORDER BY w.watchers_count DESC, w.id DESC
LIMIT 30;

-- name: GetWorkByID :one
SELECT
    id,
    title,
    title_en,
    title_kana,
    media,
    official_site_url,
    wikipedia_url,
    recommended_image_url,
    watchers_count,
    episodes_count,
    season_year,
    season_name,
    synopsis,
    created_at,
    updated_at
FROM works
WHERE id = $1;

-- name: GetWorkForArchiveByID :one
SELECT
    id,
    title,
    unpublished_at,
    deleted_at
FROM works
WHERE id = $1;

-- name: GetWorkForEditByID :one
SELECT
    id,
    title,
    title_kana,
    title_alter,
    title_en,
    title_alter_en,
    media,
    season_year,
    season_name,
    started_on,
    ended_on,
    official_site_url,
    official_site_url_en,
    wikipedia_url,
    wikipedia_url_en,
    twitter_username,
    twitter_hashtag,
    sc_tid,
    mal_anime_id,
    synopsis,
    synopsis_source,
    synopsis_en,
    synopsis_source_en,
    manual_episodes_count,
    start_episode_raw_number,
    number_format_id,
    no_episodes
FROM works
WHERE id = $1;

-- name: GetWorkForEpisodeListByID :one
-- published_episode_count and max_generatable_episode_number feed the episode list's
-- auto-generation notice. The first is how many episodes are currently published. The
-- second is the highest number the Syobocal auto-generation could assign from the work's
-- kept slots. Both aggregate other tables for one work, so they ride along with the work
-- row instead of costing the page extra round trips.
--
-- max_generatable_episode_number takes MAX, which skips NULLs, where the Rails notice reads
-- the first of the work's kept slots ordered by number descending. PostgreSQL sorts NULLs
-- first for DESC, so Rails lands on a NULL number and reports 0 as soon as the work has one
-- kept slot without a number. The label names the highest number the auto-generation can
-- reach, so keeping MAX here is deliberate rather than a gap in the port.
--
-- [Ja] published_episode_count と max_generatable_episode_number はエピソード一覧の自動生成の
-- 案内に使う。前者は作品のエピソードのうち現在公開中の件数、後者はしょぼいカレンダー由来の
-- 自動生成が作品の有効なスロットから振れる最大話数を表す。どちらも 1 作品について別テーブルを
-- 集計する値のため、作品の行と一緒に引いてページの往復を増やさない。
--
-- max_generatable_episode_number が使う MAX は NULL を飛ばすが、Rails の案内は作品の有効な
-- スロットを number 降順に並べた先頭行を読む。PostgreSQL の DESC は NULLS FIRST のため、
-- number 未設定の有効スロットが 1 件でもあれば Rails はその行に当たって 0 を報告する。
-- ラベルが名指しするのは自動生成が到達できる最大話数であり、ここで MAX を使い続けるのは
-- 移植漏れではなく意図的な選択。
SELECT
    w.id,
    w.title,
    w.no_episodes,
    w.manual_episodes_count,
    (
        SELECT COUNT(*)
        FROM episodes e
        WHERE e.work_id = w.id
            AND e.deleted_at IS NULL
            AND e.unpublished_at IS NULL
    )::bigint AS published_episode_count,
    COALESCE((
        SELECT MAX(s.number)
        FROM slots s
        WHERE s.work_id = w.id
            AND s.deleted_at IS NULL
            AND s.unpublished_at IS NULL
    ), 0)::bigint AS max_generatable_episode_number
FROM works w
WHERE w.id = $1
    AND w.deleted_at IS NULL;

-- name: GetWorkForEpisodeFormByID :one
-- The episode form needs the work to name it in the heading, drive the shared work subnav
-- and preserve the Rails manual-create guard. An editor cannot create more episodes once
-- the published count reaches manual_episodes_count, or while the work owns a slot with a
-- start time; admins still see the warning but may override it in the presentation layer.
--
-- [Ja] エピソードフォームは、見出しでの名指し、共有サブナビの出し分け、および Rails の
-- 手動作成ガードを保つために作品を取得する。公開中のエピソード数が manual_episodes_count に
-- 達した作品、または開始時刻を持つ放送枠がある作品には編集者が追加できない。管理者も警告は
-- 見るが、表示層で上書きして作成できる。
SELECT
    w.id,
    w.title,
    w.no_episodes,
    w.manual_episodes_count IS NOT NULL AND (
        SELECT COUNT(*)
        FROM episodes e
        WHERE e.work_id = w.id
            AND e.unpublished_at IS NULL
            AND e.deleted_at IS NULL
    ) >= w.manual_episodes_count AS episodes_filled,
    EXISTS (
        SELECT 1
        FROM slots s
        WHERE s.work_id = w.id
            AND s.started_at IS NOT NULL
    ) AS slots_exist
FROM works w
WHERE w.id = $1
    AND w.deleted_at IS NULL;

-- name: ExistsWorkForEpisodeCreateByID :one
-- Check the parent before parsing the submitted rows, matching the Rails action's
-- Work.without_deleted.find ordering. The authoritative row is locked and reloaded after
-- validation, so this preliminary check is not used for creation decisions.
--
-- [Ja] Rails アクションの Work.without_deleted.find と同じ順序で、送信行をパースする前に親作品を
-- 確認する。バリデーション後に正本の行をロックして再取得するため、この予備確認の結果は作成時の
-- 判断には使わない。
SELECT EXISTS (
    SELECT 1
    FROM works w
    WHERE w.id = $1
        AND w.deleted_at IS NULL
);

-- name: LockWorkForEpisodeCreateByID :one
-- Serialize bulk creates for one work before reading their numbering anchors. Keeping the
-- lock query separate from the aggregate query makes the latter run after a waiter acquires
-- the lock and therefore observe the preceding transaction's committed episodes.
--
-- [Ja] 採番の起点を読む前に、1 作品への一括作成を直列化する。ロッククエリと集計クエリを分ける
-- ことで、待機側がロックを得た後に集計を実行し、先行トランザクションがコミットしたエピソードを
-- 参照できるようにする。
SELECT w.id
FROM works w
WHERE w.id = $1
    AND w.deleted_at IS NULL
FOR UPDATE;

-- name: GetWorkForEpisodeCreateByID :one
-- episode_count and the latest_* columns anchor the sort_number the bulk create assigns: the
-- first new episode starts one step past episode_count * 100, and the episode holding the
-- greatest sort_number becomes the prev_episode_id of the first created row. Both aggregate
-- the work's episodes without filtering unpublished or deleted ones, matching the Rails form
-- (work.episodes.count), so a work whose episodes were archived does not hand out
-- sort_numbers that are already taken. anime_id decides whether the create dual-writes the
-- reference model: an episode's classification requires the parent work's anime.
--
-- latest_episode_id / latest_sort_number are 0 when the work has no episode yet. Ids are
-- positive, so the caller reads 0 as "no preceding episode" (the same shape
-- max_generatable_episode_number above uses for a work with no slot).
--
-- [Ja] episode_count と latest_* のカラムは、一括作成が振る sort_number の起点になる。最初の
-- 新規エピソードは episode_count * 100 の 1 ステップ先から始まり、sort_number が最大の
-- エピソードが最初に作る行の prev_episode_id になる。どちらも非公開・削除済みを除外せずに
-- 作品のエピソードを集計する (Rails のフォームの work.episodes.count と同じ)。エピソードを
-- 非公開にした作品で、既に使われている sort_number を振り直さないため。anime_id は作成が
-- 参照モデルへ両書きするかどうかを決める (エピソードの分類は親作品の anime を必要とする)。
--
-- latest_episode_id / latest_sort_number は、作品がまだエピソードを持たないとき 0 に
-- なる。id は正の値のため、呼び出し側は 0 を「直前のエピソードなし」と読む (上の
-- max_generatable_episode_number がスロットの無い作品に対して採るのと同じ形)。
SELECT
    w.id,
    w.anime_id,
    (
        SELECT COUNT(*)
        FROM episodes e
        WHERE e.work_id = w.id
    )::bigint AS episode_count,
    COALESCE(latest.id, 0)::bigint AS latest_episode_id,
    COALESCE(latest.sort_number, 0)::integer AS latest_sort_number,
    w.manual_episodes_count IS NOT NULL AND (
        SELECT COUNT(*)
        FROM episodes kept
        WHERE kept.work_id = w.id
            AND kept.unpublished_at IS NULL
            AND kept.deleted_at IS NULL
    ) >= w.manual_episodes_count AS episodes_filled,
    EXISTS (
        SELECT 1
        FROM slots s
        WHERE s.work_id = w.id
            AND s.started_at IS NOT NULL
    ) AS slots_exist
FROM works w
LEFT JOIN LATERAL (
    SELECT e.id, e.sort_number
    FROM episodes e
    WHERE e.work_id = w.id
    ORDER BY e.sort_number DESC, e.id DESC
    LIMIT 1
) latest ON TRUE
WHERE w.id = $1
    AND w.deleted_at IS NULL;

-- name: IncrementWorkEpisodesCount :execrows
-- Episode.create in Rails increments the published counter cache and touches its work.
-- The bulk create holds the work row lock already; add the number of newly published rows
-- atomically so the shared Rails API observes the same counter and timestamp side effects.
--
-- [Ja] Rails の Episode.create は公開話数のカウンターキャッシュを加算し、親作品を touch する。
-- 一括作成は既に作品行をロックしているため、新しく公開した行数を原子的に加算し、共有する Rails
-- API から同じカウンターとタイムスタンプの副作用が見えるようにする。
UPDATE works
SET
    episodes_count = episodes_count + sqlc.arg('created_count'),
    updated_at = NOW()
WHERE id = sqlc.arg('work_id')
    AND deleted_at IS NULL;

-- name: ListDBWorks :many
SELECT
    w.id,
    w.title,
    w.title_kana,
    w.title_en,
    w.media,
    w.sc_tid,
    w.mal_anime_id,
    w.season_year,
    w.season_name,
    w.watchers_count,
    w.unpublished_at,
    w.deleted_at,
    wi.image_data
FROM works w
LEFT JOIN work_images wi ON w.id = wi.work_id
WHERE w.deleted_at IS NULL
    AND (sqlc.narg('filter_no_episodes')::boolean IS NOT TRUE OR (
        w.no_episodes = false AND NOT EXISTS (
            SELECT 1 FROM episodes e
            WHERE e.work_id = w.id AND e.deleted_at IS NULL AND e.unpublished_at IS NULL
        )
    ))
    AND (sqlc.narg('filter_no_image')::boolean IS NOT TRUE OR wi.id IS NULL)
    AND (sqlc.narg('filter_no_season')::boolean IS NOT TRUE OR (w.season_year IS NULL AND w.season_name IS NULL))
    AND (sqlc.narg('filter_no_slots')::boolean IS NOT TRUE OR NOT EXISTS (
        SELECT 1 FROM slots s
        WHERE s.work_id = w.id AND s.deleted_at IS NULL AND s.unpublished_at IS NULL
    ))
    AND (sqlc.narg('season_year')::int IS NULL OR w.season_year = sqlc.narg('season_year'))
    AND (sqlc.narg('season_name')::int IS NULL OR w.season_name = sqlc.narg('season_name'))
    AND (
        coalesce(cardinality(sqlc.arg('season_years')::int[]), 0) = 0
        OR EXISTS (
            SELECT 1
            FROM generate_subscripts(sqlc.arg('season_years')::int[], 1) AS i
            WHERE w.season_year = (sqlc.arg('season_years')::int[])[i]
                AND w.season_name = (sqlc.arg('season_names')::int[])[i]
        )
    )
ORDER BY w.id DESC
LIMIT sqlc.arg('per_page')
OFFSET sqlc.arg('page_offset')::bigint;

-- name: CountDBWorks :one
SELECT COUNT(*)
FROM works w
LEFT JOIN work_images wi ON w.id = wi.work_id
WHERE w.deleted_at IS NULL
    AND (sqlc.narg('filter_no_episodes')::boolean IS NOT TRUE OR (
        w.no_episodes = false AND NOT EXISTS (
            SELECT 1 FROM episodes e
            WHERE e.work_id = w.id AND e.deleted_at IS NULL AND e.unpublished_at IS NULL
        )
    ))
    AND (sqlc.narg('filter_no_image')::boolean IS NOT TRUE OR wi.id IS NULL)
    AND (sqlc.narg('filter_no_season')::boolean IS NOT TRUE OR (w.season_year IS NULL AND w.season_name IS NULL))
    AND (sqlc.narg('filter_no_slots')::boolean IS NOT TRUE OR NOT EXISTS (
        SELECT 1 FROM slots s
        WHERE s.work_id = w.id AND s.deleted_at IS NULL AND s.unpublished_at IS NULL
    ))
    AND (sqlc.narg('season_year')::int IS NULL OR w.season_year = sqlc.narg('season_year'))
    AND (sqlc.narg('season_name')::int IS NULL OR w.season_name = sqlc.narg('season_name'))
    AND (
        coalesce(cardinality(sqlc.arg('season_years')::int[]), 0) = 0
        OR EXISTS (
            SELECT 1
            FROM generate_subscripts(sqlc.arg('season_years')::int[], 1) AS i
            WHERE w.season_year = (sqlc.arg('season_years')::int[])[i]
                AND w.season_name = (sqlc.arg('season_names')::int[])[i]
        )
    );

-- name: ListWorksForAnimeSyncByIDs :many
SELECT
    id,
    title,
    title_kana,
    title_ro,
    title_en,
    title_alter,
    title_alter_en,
    media,
    synopsis,
    synopsis_en,
    synopsis_source,
    synopsis_source_en,
    unpublished_at,
    deleted_at,
    no_episodes,
    manual_episodes_count,
    start_episode_raw_number,
    number_format_id,
    anime_id
FROM works
WHERE id = ANY($1::bigint[])
ORDER BY id;

-- name: ListWorksForSatelliteSyncByIDs :many
SELECT
    id,
    anime_id,
    sc_tid,
    mal_anime_id,
    official_site_url,
    official_site_url_en,
    wikipedia_url,
    wikipedia_url_en,
    twitter_username,
    twitter_hashtag,
    season_year,
    season_name,
    started_on,
    ended_on
FROM works
WHERE id = ANY($1::bigint[])
ORDER BY id;

-- name: ListWorkIDsAfter :many
SELECT id
FROM works
WHERE id > sqlc.arg('after_id')
ORDER BY id
LIMIT sqlc.arg('batch_size');

-- name: UpdateWorkAnimeID :exec
UPDATE works
SET anime_id = $2
WHERE id = $1;

-- name: UpdateWorkUnpublishedAt :exec
UPDATE works
SET
    unpublished_at = sqlc.narg('unpublished_at'),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: UpdateWorkDeletedAt :exec
UPDATE works
SET
    deleted_at = sqlc.narg('deleted_at'),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: UpdateWork :exec
UPDATE works
SET
    title = sqlc.arg('title'),
    title_kana = sqlc.arg('title_kana'),
    title_alter = sqlc.arg('title_alter'),
    title_en = sqlc.arg('title_en'),
    title_alter_en = sqlc.arg('title_alter_en'),
    media = sqlc.arg('media'),
    season_year = sqlc.narg('season_year'),
    season_name = sqlc.narg('season_name'),
    started_on = sqlc.narg('started_on'),
    ended_on = sqlc.narg('ended_on'),
    official_site_url = sqlc.arg('official_site_url'),
    official_site_url_en = sqlc.arg('official_site_url_en'),
    wikipedia_url = sqlc.arg('wikipedia_url'),
    wikipedia_url_en = sqlc.arg('wikipedia_url_en'),
    twitter_username = sqlc.narg('twitter_username'),
    twitter_hashtag = sqlc.narg('twitter_hashtag'),
    sc_tid = sqlc.narg('sc_tid'),
    mal_anime_id = sqlc.narg('mal_anime_id'),
    synopsis = sqlc.arg('synopsis'),
    synopsis_source = sqlc.arg('synopsis_source'),
    synopsis_en = sqlc.arg('synopsis_en'),
    synopsis_source_en = sqlc.arg('synopsis_source_en'),
    manual_episodes_count = sqlc.narg('manual_episodes_count'),
    start_episode_raw_number = sqlc.arg('start_episode_raw_number'),
    number_format_id = sqlc.narg('number_format_id'),
    no_episodes = sqlc.arg('no_episodes'),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: CreateWork :one
INSERT INTO works (
    title,
    title_kana,
    title_alter,
    title_en,
    title_alter_en,
    media,
    season_year,
    season_name,
    started_on,
    ended_on,
    official_site_url,
    official_site_url_en,
    wikipedia_url,
    wikipedia_url_en,
    twitter_username,
    twitter_hashtag,
    sc_tid,
    mal_anime_id,
    synopsis,
    synopsis_source,
    synopsis_en,
    synopsis_source_en,
    manual_episodes_count,
    start_episode_raw_number,
    number_format_id,
    no_episodes,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    sqlc.narg('season_year'),
    sqlc.narg('season_name'),
    sqlc.narg('started_on'),
    sqlc.narg('ended_on'),
    $7, $8, $9, $10,
    sqlc.narg('twitter_username'),
    sqlc.narg('twitter_hashtag'),
    sqlc.narg('sc_tid'),
    sqlc.narg('mal_anime_id'),
    $11, $12, $13, $14,
    sqlc.narg('manual_episodes_count'),
    $15,
    sqlc.narg('number_format_id'),
    $16,
    NOW(),
    NOW()
) RETURNING id;
-- name: ExistsKeptWorkByTitle :one
SELECT EXISTS (
    SELECT 1
    FROM works
    WHERE title = sqlc.arg('title')
        AND deleted_at IS NULL
        AND unpublished_at IS NULL
        AND (sqlc.narg('exclude_id')::bigint IS NULL OR id <> sqlc.narg('exclude_id'))
);
