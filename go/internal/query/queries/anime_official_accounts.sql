-- name: CreateAnimeOfficialAccount :one
INSERT INTO anime_official_accounts (
    anime_id,
    service,
    account,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, NOW(), NOW()
) RETURNING *;

-- name: ListAnimeOfficialAccountsByAnimeIDs :many
SELECT * FROM anime_official_accounts
WHERE anime_id = ANY($1::bigint[])
ORDER BY anime_id, service;

-- name: UpdateAnimeOfficialAccount :exec
UPDATE anime_official_accounts
SET
    account = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteAnimeOfficialAccount :exec
DELETE FROM anime_official_accounts
WHERE id = $1;
