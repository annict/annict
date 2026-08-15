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

-- name: UpsertAnimeClassification :exec
-- Recreate a missing classification or update the existing row atomically. Episode editing
-- uses this instead of a preceding existence check so a concurrent delete cannot leave the
-- dual-written anime without its classification.
--
-- [Ja] 欠損した分類の再作成と既存行の更新をアトミックに行う。エピソード編集では事前の
-- 存在確認をせずこのクエリを使い、並行削除によって両書き先の anime だけが分類なしで残るのを
-- 防ぐ。
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
)
ON CONFLICT (anime_id) DO UPDATE
SET
    kind = EXCLUDED.kind,
    parent_anime_id = EXCLUDED.parent_anime_id,
    number = EXCLUDED.number,
    number_text = EXCLUDED.number_text,
    sort_number = EXCLUDED.sort_number,
    standalone = EXCLUDED.standalone,
    number_format_id = EXCLUDED.number_format_id,
    episode_start_number = EXCLUDED.episode_start_number,
    expected_episodes_count = EXCLUDED.expected_episodes_count,
    updated_at = NOW();

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
