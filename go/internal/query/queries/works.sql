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

-- name: GetWorkForStateChangeByID :one
-- Annict DBの作品の状態変更 (非公開・公開・削除) の確認画面が共有する射影。各画面が対象を
-- 名指しするタイトルと、呼び出し側が現在の状態を導出してその操作を適用できない作品を弾くための
-- 作品状態のsourceを運ぶ。3つの画面が受け付ける状態はそれぞれ異なるため、ここでは状態を
-- 絞り込まない。
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
    no_episodes,
    updated_at
FROM works
WHERE id = $1;

-- name: GetWorkForEpisodeListByID :one
-- published_episode_countとmax_generatable_episode_numberはエピソード一覧の自動生成の
-- 案内に使う。前者は作品のエピソードのうち現在公開中の件数、後者は自動生成が作品の放送枠から
-- 作るエピソードの数を表す。どちらも1作品について別テーブルを集計する値のため、作品の行と
-- 一緒に引いてページの往復を増やさない。
--
-- max_generatable_episode_numberは、自動生成の対象と同じ条件 (未削除・公開中で、再放送でも
-- 不定期放送でもなく、番組と回数を持つ) の放送枠について、作品内の順番 (回数から、その
-- 放送枠より前にある番組の不定期放送の数を引いたもの) の最大値をとる。自動生成が話数の計算と
-- 予定エピソード数との比較に使うのと同じ数え方で、不定期放送のある番組でも回数そのものより
-- 実際に作られる数を表す。
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
        SELECT MAX(s.number - (
            SELECT COUNT(*)
            FROM slots i
            WHERE i.program_id = s.program_id
                AND i.irregular
                AND i.deleted_at IS NULL
                AND i.unpublished_at IS NULL
                AND i.number >= p.minimum_episode_generatable_number
                AND i.number < s.number
        ))
        FROM slots s
        JOIN programs p ON p.id = s.program_id
        WHERE s.work_id = w.id
            AND s.deleted_at IS NULL
            AND s.unpublished_at IS NULL
            AND NOT s.rebroadcast
            AND NOT s.irregular
            AND s.number IS NOT NULL
    ), 0)::bigint AS max_generatable_episode_number
FROM works w
WHERE w.id = $1
    AND w.deleted_at IS NULL;

-- name: GetWorkForEpisodeFormByID :one
-- エピソードフォームは、見出しでの名指し、共有サブナビの出し分け、およびRailsの
-- 手動作成ガードを保つために作品を取得する。公開中のエピソード数がmanual_episodes_countに
-- 達した作品には編集者が追加できない。管理者も警告は見るが、表示層で上書きして作成できる。
--
-- has_active_slotsは、作品が有効な放送枠 (未削除・公開中) を持つか。持つ作品では自動生成が
-- 話数でエピソードを照合するため、一括作成のすべての行に話数を求める。
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
            AND s.deleted_at IS NULL
            AND s.unpublished_at IS NULL
    ) AS has_active_slots
FROM works w
WHERE w.id = $1
    AND w.deleted_at IS NULL;

-- name: LockWorkForEpisodeCreateByID :one
-- 採番の起点を読む前に、1作品への一括作成を直列化する。ロッククエリと集計クエリを分ける
-- ことで、待機側がロックを得た後に集計を実行し、先行トランザクションがコミットしたエピソードを
-- 参照できるようにする。
SELECT w.id
FROM works w
WHERE w.id = $1
    AND w.deleted_at IS NULL
FOR UPDATE;

-- name: GetWorkForEpisodeCreateByID :one
-- latest_* のカラムは、一括作成が振るsort_numberの起点になる。最初の新規エピソードは
-- 作品の最大のsort_numberの1ステップ先から始まり、その最大の行が最初に作る行の
-- prev_episode_idになる。非公開・削除済みを除外せずに集計する。削除済みの行が持つ値と
-- ぶつけないためと、開始話数が1でない作品で自動生成が置いたエピソードより前に並ばないように
-- するため。anime_idは作成が参照モデルへ両書きするかどうかを決める (エピソードの分類は親作品の
-- animeを必要とする)。
--
-- latest_episode_id / latest_sort_numberは、作品がまだエピソードを持たないとき0に
-- なる。idは正の値のため、呼び出し側は0を「直前のエピソードなし」と読む (上の
-- max_generatable_episode_numberがスロットの無い作品に対して採るのと同じ形)。
SELECT
    w.id,
    w.anime_id,
    COALESCE(latest.id, 0)::bigint AS latest_episode_id,
    COALESCE(latest.sort_number, 0)::integer AS latest_sort_number,
    EXISTS (
        SELECT 1
        FROM slots s
        WHERE s.work_id = w.id
            AND s.deleted_at IS NULL
            AND s.unpublished_at IS NULL
    ) AS has_active_slots,
    w.manual_episodes_count IS NOT NULL AND (
        SELECT COUNT(*)
        FROM episodes kept
        WHERE kept.work_id = w.id
            AND kept.unpublished_at IS NULL
            AND kept.deleted_at IS NULL
    ) >= w.manual_episodes_count AS episodes_filled
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
-- RailsのEpisode.createは公開話数のカウンターキャッシュを加算し、親作品をtouchする。
-- 一括作成は既に作品行をロックしているため、新しく公開した行数を原子的に加算し、共有するRails
-- APIから同じカウンターとタイムスタンプの副作用が見えるようにする。
UPDATE works
SET
    episodes_count = episodes_count + sqlc.arg('created_count'),
    updated_at = NOW()
WHERE id = sqlc.arg('work_id')
    AND deleted_at IS NULL;

-- name: GetWorkForEpisodeGenerationByID :one
-- 放送枠からエピソードを作るときに、作品をロックした後で読む。話数の計算 (開始話数)・予定
-- エピソード数との比較・表示の番号の書式・作品のanimeを返す。max_sort_numberは削除済み・
-- 非公開を含む全エピソードの最大値で、末尾に足すエピソードの並び順の起点になる。削除済みの行が
-- 持つ値と並び順がぶつからないようにするため。作品がエピソードを持たないあいだは0になる。
-- 一覧を読んだ後に非公開・削除にされた作品は返さず、呼び出し側はその放送枠を飛ばす。
SELECT
    w.id,
    w.title,
    w.anime_id,
    w.start_episode_raw_number,
    w.manual_episodes_count,
    nf.id AS number_format_id,
    COALESCE(nf.format, '')::varchar AS number_format_format,
    COALESCE(nf.data, '{}')::varchar[] AS number_format_data,
    COALESCE((
        SELECT MAX(e.sort_number)
        FROM episodes e
        WHERE e.work_id = w.id
    ), 0)::integer AS max_sort_number
FROM works w
LEFT JOIN number_formats nf ON nf.id = w.number_format_id
WHERE w.id = $1
    AND w.deleted_at IS NULL
    AND w.unpublished_at IS NULL;

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
    AND (
        coalesce(cardinality(sqlc.arg('season_years')::int[]), 0) = 0
        OR EXISTS (
            SELECT 1
            FROM generate_subscripts(sqlc.arg('season_years')::int[], 1) AS i
            WHERE w.season_year = (sqlc.arg('season_years')::int[])[i]
                -- 季節の0は「季節未登録」を表し、season_nameがNULLの作品に一致させる
                AND w.season_name IS NOT DISTINCT FROM NULLIF((sqlc.arg('season_names')::int[])[i], 0)
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
    AND (
        coalesce(cardinality(sqlc.arg('season_years')::int[]), 0) = 0
        OR EXISTS (
            SELECT 1
            FROM generate_subscripts(sqlc.arg('season_years')::int[], 1) AS i
            WHERE w.season_year = (sqlc.arg('season_years')::int[])[i]
                -- 季節の0は「季節未登録」を表し、season_nameがNULLの作品に一致させる
                AND w.season_name IS NOT DISTINCT FROM NULLIF((sqlc.arg('season_names')::int[])[i], 0)
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

-- name: UpdateWork :execrows
-- 送信が名乗る版の照合は、直前の読み取りとの比較ではなくUPDATEの中で行う。比較と書き込み
-- の間に他の書き込みが挟まらないようにするため。共有カラムはNULL許容のためNULLも明示的な版
-- であり、IS NOT DISTINCT FROMなら2文に分けずに照合できる。書き込みはupdated_atを進めるので、
-- 同じNULL版からの2件目はもう一致しない。したがって1行も更新されないことは、その間に他者が
-- 行を書いたことを意味し、呼び出し側は上書きせず競合として報告する。
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
WHERE id = sqlc.arg('id')
    AND updated_at IS NOT DISTINCT FROM sqlc.narg('version')::timestamptz;

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
