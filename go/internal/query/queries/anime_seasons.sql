-- name: CreateAnimeSeason :one
INSERT INTO anime_seasons (
    anime_id,
    year,
    name,
    is_primary,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, NOW(), NOW()
) RETURNING *;

-- name: ListAnimeSeasonsByAnimeIDs :many
SELECT * FROM anime_seasons
WHERE anime_id = ANY($1::bigint[])
ORDER BY anime_id, year, name;

-- name: DeleteAnimeSeason :exec
DELETE FROM anime_seasons
WHERE id = $1;
