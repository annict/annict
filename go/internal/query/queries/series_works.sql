-- name: ReconcileSeriesWorkAnimeIDsAfter :one
-- animes同期バッチの関連テーブルの照合。idがafter_idより大きいseries_worksの1ページについて、
-- anime_idをwork_idの指す作品のanime_idと照合し、違う行だけを書き換える。anime_idはNOT NULLのため、
-- 作品のanime_idがまだ無い行があると、NULLを書こうとして制約違反で失敗する。updated_atは編集者が
-- 行を変えた時刻のため変えない。
-- last_idはページの最後の行のid (ページが空なら0) で、次のページのafter_idに使う。
WITH page AS (
    SELECT series_works.id, works.anime_id
    FROM series_works
    LEFT JOIN works ON works.id = series_works.work_id
    WHERE series_works.id > sqlc.arg('after_id')
    ORDER BY series_works.id
    LIMIT sqlc.arg('batch_size')
), updated AS (
    UPDATE series_works
    SET anime_id = page.anime_id
    FROM page
    WHERE series_works.id = page.id
        AND series_works.anime_id IS DISTINCT FROM page.anime_id
    RETURNING series_works.id
)
SELECT
    COALESCE((SELECT MAX(page.id) FROM page), 0)::bigint AS last_id,
    (SELECT COUNT(*) FROM updated)::bigint AS updated_count;
