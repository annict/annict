-- name: GetCastsByWorkIDs :many
-- キャスト自身が公開状態 (unpublished_atとdeleted_atがともにNULL) で、人物が公開中 (published)
-- の行だけを返す。非公開 (archived)・掲載停止 (delisted)・削除 (deleted) の人物の行は、Rails版の
-- 作品ページと同じく除外する
SELECT
    c.id,
    c.work_id,
    c.name,
    c.name_en,
    c.sort_number,
    c.person_id,
    ch.id as character_id,
    ch.name as character_name,
    ch.name_en as character_name_en,
    p.name as person_name,
    p.name_en as person_name_en
FROM casts c
LEFT JOIN characters ch ON c.character_id = ch.id
INNER JOIN people p ON c.person_id = p.id
WHERE c.work_id = ANY($1::bigint[])
  AND c.deleted_at IS NULL
  AND c.unpublished_at IS NULL
  AND p.status = 'published'
ORDER BY c.work_id, c.sort_number;

-- name: ReconcileCastAnimeIDsAfter :one
-- animes同期バッチの関連テーブルの照合。idがafter_idより大きいcastsの1ページについて、
-- anime_idをwork_idの指す作品のanime_idと照合し、違う行だけを書き換える。work_idがNULLの行と、
-- 作品のanime_idがまだ無い行はNULLにする。updated_atは編集者が行を変えた時刻のため変えない。
-- last_idはページの最後の行のid (ページが空なら0) で、次のページのafter_idに使う。
WITH page AS (
    SELECT casts.id, works.anime_id
    FROM casts
    LEFT JOIN works ON works.id = casts.work_id
    WHERE casts.id > sqlc.arg('after_id')
    ORDER BY casts.id
    LIMIT sqlc.arg('batch_size')
), updated AS (
    UPDATE casts
    SET anime_id = page.anime_id
    FROM page
    WHERE casts.id = page.id
        AND casts.anime_id IS DISTINCT FROM page.anime_id
    RETURNING casts.id
)
SELECT
    COALESCE((SELECT MAX(page.id) FROM page), 0)::bigint AS last_id,
    (SELECT COUNT(*) FROM updated)::bigint AS updated_count;
