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

-- name: ListAnimeIDsBySeasons :many
-- 年か季節に一致する季節の行を持つanimeのIDを返す。代表の季節に限らず、どの行に一致してもよい。
-- seasonsは `2026-spring` の形の季節の配列で、季節の分からない行は `2026-unknown` で表す。
-- season_yearsはseasonsの年の配列で、年のインデックスで行を絞るために渡す
SELECT DISTINCT anime_id
FROM anime_seasons
WHERE (sqlc.narg(year)::integer IS NULL OR year = sqlc.narg(year)::integer)
    AND (
        sqlc.narg(seasons)::text[] IS NULL
        OR (
            year = ANY(sqlc.narg(season_years)::integer[])
            AND year::text || '-' || COALESCE(name::text, 'unknown') = ANY(sqlc.narg(seasons)::text[])
        )
    );
