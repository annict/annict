-- name: CreateAnimeEvent :one
INSERT INTO anime_events (
    anime_id,
    kind,
    started_on,
    ended_on,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, NOW(), NOW()
) RETURNING *;

-- name: ListAnimeEventsByAnimeIDs :many
SELECT * FROM anime_events
WHERE anime_id = ANY($1::bigint[])
ORDER BY anime_id, kind;

-- name: UpdateAnimeEvent :exec
UPDATE anime_events
SET
    started_on = $2,
    ended_on = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteAnimeEvent :exec
DELETE FROM anime_events
WHERE id = $1;
