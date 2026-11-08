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
-- 欠損した分類の再作成と既存行の更新をアトミックに行う。エピソード編集では事前の
-- 存在確認をせずこのクエリを使い、並行削除によって両書き先のanimeだけが分類なしで残るのを
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

-- name: CountPublishedEpisodesByParentAnimeIDs :many
-- 作品ごとの公開中のエピソードの数を返す。エピソードの無い作品の行は返さない
SELECT ac.parent_anime_id::bigint AS parent_anime_id, COUNT(*) AS episodes_count
FROM anime_classifications ac
INNER JOIN animes a ON a.id = ac.anime_id
WHERE ac.parent_anime_id = ANY($1::bigint[])
    AND a.status = 'published'
GROUP BY ac.parent_anime_id;

-- name: ListPublishedEpisodeNeighbors :many
-- 指定したエピソードについて、同じ作品の公開中のエピソードを (sort_number, anime_id) の順に並べたときの前後を返す。
-- 前後を作品ごとにまとめて求めるため、親の作品の公開中のエピソードをすべて並べてから絞る。
-- 前後が無いときは0を返す (animes.idは1から始まる)
SELECT
    x.anime_id::bigint AS anime_id,
    COALESCE(x.prev_anime_id, 0)::bigint AS prev_anime_id,
    COALESCE(x.next_anime_id, 0)::bigint AS next_anime_id
FROM (
    SELECT
        ac.anime_id,
        LAG(ac.anime_id) OVER w AS prev_anime_id,
        LEAD(ac.anime_id) OVER w AS next_anime_id
    FROM anime_classifications ac
    INNER JOIN animes a ON a.id = ac.anime_id
    WHERE ac.parent_anime_id = ANY(sqlc.arg(parent_anime_ids)::bigint[])
        AND a.status = 'published'
    WINDOW w AS (PARTITION BY ac.parent_anime_id ORDER BY ac.sort_number, ac.anime_id)
) x
WHERE x.anime_id = ANY(sqlc.arg(anime_ids)::bigint[]);
