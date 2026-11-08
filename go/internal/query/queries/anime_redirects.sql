-- name: GetAnimeRedirectByOldAnimeID :one
SELECT * FROM anime_redirects
WHERE old_anime_id = $1
LIMIT 1;

-- name: ListAnimeRedirectsByOldAnimeIDs :many
SELECT * FROM anime_redirects
WHERE old_anime_id = ANY(sqlc.arg(old_anime_ids)::bigint[]);
