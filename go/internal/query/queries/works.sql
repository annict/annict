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

-- name: GetCastsByWorkIDs :many
SELECT
    c.id,
    c.work_id,
    c.name,
    c.name_en,
    c.sort_number,
    ch.name as character_name,
    ch.name_en as character_name_en,
    p.name as person_name,
    p.name_en as person_name_en
FROM casts c
LEFT JOIN characters ch ON c.character_id = ch.id
LEFT JOIN people p ON c.person_id = p.id
WHERE c.work_id = ANY($1::bigint[])
ORDER BY c.work_id, c.sort_number;

-- name: GetStaffsByWorkIDs :many
SELECT
    s.id,
    s.work_id,
    s.name,
    s.name_en,
    s.role,
    s.role_other,
    s.role_other_en,
    s.sort_number
FROM staffs s
WHERE s.work_id = ANY($1::bigint[])
  AND s.role != 'other'
ORDER BY s.work_id, s.sort_number;

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