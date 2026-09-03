-- name: CreateAnimeHashtag :one
INSERT INTO anime_hashtags (
    anime_id,
    hashtag,
    created_at,
    updated_at
) VALUES (
    $1, $2, NOW(), NOW()
) RETURNING *;

-- name: ListAnimeHashtagsByAnimeIDs :many
SELECT * FROM anime_hashtags
WHERE anime_id = ANY($1::bigint[])
ORDER BY anime_id, sort_number, hashtag;

-- name: DeleteAnimeHashtag :exec
DELETE FROM anime_hashtags
WHERE id = $1;
