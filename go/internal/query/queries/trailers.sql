-- name: ReconcileTrailerAnimeIDsAfter :one
-- animes同期バッチの関連テーブルの照合。idがafter_idより大きいtrailersの1ページについて、
-- anime_idをwork_idの指す作品のanime_idと照合し、違う行だけを書き換える。anime_idはNOT NULLのため、
-- 作品のanime_idがまだ無い行があると、NULLを書こうとして制約違反で失敗する。updated_atは編集者が
-- 行を変えた時刻のため変えない。
-- last_idはページの最後の行のid (ページが空なら0) で、次のページのafter_idに使う。
WITH page AS (
    SELECT trailers.id, works.anime_id
    FROM trailers
    LEFT JOIN works ON works.id = trailers.work_id
    WHERE trailers.id > sqlc.arg('after_id')
    ORDER BY trailers.id
    LIMIT sqlc.arg('batch_size')
), updated AS (
    UPDATE trailers
    SET anime_id = page.anime_id
    FROM page
    WHERE trailers.id = page.id
        AND trailers.anime_id IS DISTINCT FROM page.anime_id
    RETURNING trailers.id
)
SELECT
    COALESCE((SELECT MAX(page.id) FROM page), 0)::bigint AS last_id,
    (SELECT COUNT(*) FROM updated)::bigint AS updated_count;
