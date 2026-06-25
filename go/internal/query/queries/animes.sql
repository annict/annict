-- name: CreateAnime :one
INSERT INTO animes (
    title,
    title_kana,
    title_ro,
    title_en,
    title_alter,
    title_alter_ro,
    title_alter_en,
    title_alter_other,
    media,
    release_status,
    synopsis,
    synopsis_en,
    synopsis_source,
    synopsis_source_en,
    status,
    archive_message,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW()
) RETURNING *;

-- name: GetAnimeByID :one
SELECT * FROM animes
WHERE id = $1
LIMIT 1;

-- name: ListAnimesByIDs :many
SELECT * FROM animes
WHERE id = ANY($1::bigint[])
ORDER BY id;

-- name: UpdateAnime :exec
UPDATE animes
SET
    title = $2,
    title_kana = $3,
    title_ro = $4,
    title_en = $5,
    title_alter = $6,
    title_alter_ro = $7,
    title_alter_en = $8,
    title_alter_other = $9,
    media = $10,
    release_status = $11,
    synopsis = $12,
    synopsis_en = $13,
    synopsis_source = $14,
    synopsis_source_en = $15,
    status = $16,
    archive_message = $17,
    updated_at = NOW()
WHERE id = $1;
