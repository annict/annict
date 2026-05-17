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

-- name: ListDBWorks :many
SELECT
    w.id,
    w.title,
    w.season_year,
    w.season_name,
    w.watchers_count,
    w.status,
    CASE WHEN wi.id IS NOT NULL THEN true ELSE false END AS has_image
FROM works w
LEFT JOIN work_images wi ON w.id = wi.work_id
WHERE w.status != 'deleted'
    AND (sqlc.narg('filter_no_episodes')::boolean IS NOT TRUE OR (
        w.no_episodes = false AND NOT EXISTS (
            SELECT 1 FROM episodes e WHERE e.work_id = w.id AND e.status = 'published'
        )
    ))
    AND (sqlc.narg('filter_no_image')::boolean IS NOT TRUE OR wi.id IS NULL)
    AND (sqlc.narg('filter_no_season')::boolean IS NOT TRUE OR (w.season_year IS NULL AND w.season_name IS NULL))
    AND (sqlc.narg('season_year')::int IS NULL OR w.season_year = sqlc.narg('season_year'))
    AND (sqlc.narg('season_name')::int IS NULL OR w.season_name = sqlc.narg('season_name'))
ORDER BY w.id DESC
LIMIT sqlc.arg('per_page')
OFFSET sqlc.arg('page_offset');

-- name: CountDBWorks :one
SELECT COUNT(*)
FROM works w
LEFT JOIN work_images wi ON w.id = wi.work_id
WHERE w.status != 'deleted'
    AND (sqlc.narg('filter_no_episodes')::boolean IS NOT TRUE OR (
        w.no_episodes = false AND NOT EXISTS (
            SELECT 1 FROM episodes e WHERE e.work_id = w.id AND e.status = 'published'
        )
    ))
    AND (sqlc.narg('filter_no_image')::boolean IS NOT TRUE OR wi.id IS NULL)
    AND (sqlc.narg('filter_no_season')::boolean IS NOT TRUE OR (w.season_year IS NULL AND w.season_name IS NULL))
    AND (sqlc.narg('season_year')::int IS NULL OR w.season_year = sqlc.narg('season_year'))
    AND (sqlc.narg('season_name')::int IS NULL OR w.season_name = sqlc.narg('season_name'));

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