-- name: CreateAnimeLink :one
INSERT INTO anime_links (
    anime_id,
    kind,
    language,
    url,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, NOW(), NOW()
) RETURNING *;

-- name: ListAnimeLinksByAnimeIDs :many
SELECT * FROM anime_links
WHERE anime_id = ANY($1::bigint[])
ORDER BY anime_id, kind, language;

-- name: UpdateAnimeLink :exec
UPDATE anime_links
SET
    url = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteAnimeLink :exec
DELETE FROM anime_links
WHERE id = $1;
