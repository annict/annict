-- name: ReconcileVodTitleAnimeIDsAfter :one
-- animes同期バッチの関連テーブルの照合。idがafter_idより大きいvod_titlesの1ページについて、
-- anime_idをwork_idの指す作品のanime_idと照合し、違う行だけを書き換える。work_idがNULLの行と、
-- 作品のanime_idがまだ無い行はNULLにする。updated_atは編集者が行を変えた時刻のため変えない。
-- last_idはページの最後の行のid (ページが空なら0) で、次のページのafter_idに使う。
WITH page AS (
    SELECT vod_titles.id, works.anime_id
    FROM vod_titles
    LEFT JOIN works ON works.id = vod_titles.work_id
    WHERE vod_titles.id > sqlc.arg('after_id')
    ORDER BY vod_titles.id
    LIMIT sqlc.arg('batch_size')
), updated AS (
    UPDATE vod_titles
    SET anime_id = page.anime_id
    FROM page
    WHERE vod_titles.id = page.id
        AND vod_titles.anime_id IS DISTINCT FROM page.anime_id
    RETURNING vod_titles.id
)
SELECT
    COALESCE((SELECT MAX(page.id) FROM page), 0)::bigint AS last_id,
    (SELECT COUNT(*) FROM updated)::bigint AS updated_count;
