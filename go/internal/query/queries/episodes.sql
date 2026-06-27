-- name: ListEpisodesForAnimeSyncByIDs :many
SELECT
    e.id,
    e.work_id,
    e.title,
    e.title_ro,
    e.title_en,
    e.number,
    e.sort_number,
    e.raw_number,
    e.status,
    e.archive_message,
    e.anime_id,
    w.anime_id AS parent_anime_id
FROM episodes e
JOIN works w ON e.work_id = w.id
WHERE e.id = ANY($1::bigint[])
ORDER BY e.id;

-- name: ListEpisodeIDsAfter :many
SELECT id
FROM episodes
WHERE id > sqlc.arg('after_id')
ORDER BY id
LIMIT sqlc.arg('batch_size');

-- name: UpdateEpisodeAnimeID :exec
UPDATE episodes
SET anime_id = $2
WHERE id = $1;
