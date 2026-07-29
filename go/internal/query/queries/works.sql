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
            SELECT 1 FROM episodes e WHERE e.work_id = w.id AND e.status = 'published'
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
OFFSET sqlc.arg('page_offset');

-- name: CountDBWorks :one
SELECT COUNT(*)
FROM works w
LEFT JOIN work_images wi ON w.id = wi.work_id
WHERE w.deleted_at IS NULL
    AND (sqlc.narg('filter_no_episodes')::boolean IS NOT TRUE OR (
        w.no_episodes = false AND NOT EXISTS (
            SELECT 1 FROM episodes e WHERE e.work_id = w.id AND e.status = 'published'
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
