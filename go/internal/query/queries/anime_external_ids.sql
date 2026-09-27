-- name: CreateAnimeExternalID :one
INSERT INTO anime_external_ids (
    anime_id,
    service,
    external_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, NOW(), NOW()
) RETURNING *;

-- name: ListAnimeExternalIDsByAnimeIDs :many
SELECT * FROM anime_external_ids
WHERE anime_id = ANY($1::bigint[])
ORDER BY anime_id, service;

-- name: UpdateAnimeExternalID :exec
UPDATE anime_external_ids
SET
    external_id = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteAnimeExternalID :exec
DELETE FROM anime_external_ids
WHERE id = $1;
