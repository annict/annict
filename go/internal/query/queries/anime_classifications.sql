-- name: CreateAnimeClassification :one
INSERT INTO anime_classifications (
    anime_id,
    kind,
    parent_anime_id,
    number,
    number_text,
    sort_number,
    standalone,
    number_format_id,
    episode_start_number,
    expected_episodes_count,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
) RETURNING *;

-- name: GetAnimeClassificationByAnimeID :one
SELECT * FROM anime_classifications
WHERE anime_id = $1
LIMIT 1;

-- name: ListAnimeClassificationsByAnimeIDs :many
SELECT * FROM anime_classifications
WHERE anime_id = ANY($1::bigint[])
ORDER BY anime_id;

-- name: UpdateAnimeClassificationByAnimeID :exec
UPDATE anime_classifications
SET
    kind = $2,
    parent_anime_id = $3,
    number = $4,
    number_text = $5,
    sort_number = $6,
    standalone = $7,
    number_format_id = $8,
    episode_start_number = $9,
    expected_episodes_count = $10,
    updated_at = NOW()
WHERE anime_id = $1;
