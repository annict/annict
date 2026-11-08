-- name: CreateWorkImage :one
INSERT INTO work_images (
    work_id,
    user_id,
    image_data,
    copyright,
    asin,
    color_rgb,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id;

-- name: ReconcileWorkImageAnimeIDsAfter :one
-- animes同期バッチの関連テーブルの照合。idがafter_idより大きいwork_imagesの1ページについて、
-- anime_idをwork_idの指す作品のanime_idと照合し、違う行だけを書き換える。work_idがNULLの行と、
-- 作品のanime_idがまだ無い行はNULLにする。updated_atは編集者が行を変えた時刻のため変えない。
-- last_idはページの最後の行のid (ページが空なら0) で、次のページのafter_idに使う。
WITH page AS (
    SELECT work_images.id, works.anime_id
    FROM work_images
    LEFT JOIN works ON works.id = work_images.work_id
    WHERE work_images.id > sqlc.arg('after_id')
    ORDER BY work_images.id
    LIMIT sqlc.arg('batch_size')
), updated AS (
    UPDATE work_images
    SET anime_id = page.anime_id
    FROM page
    WHERE work_images.id = page.id
        AND work_images.anime_id IS DISTINCT FROM page.anime_id
    RETURNING work_images.id
)
SELECT
    COALESCE((SELECT MAX(page.id) FROM page), 0)::bigint AS last_id,
    (SELECT COUNT(*) FROM updated)::bigint AS updated_count;
