-- name: ReconcileProgramAnimeIDsAfter :one
-- animes同期バッチの関連テーブルの照合。idがafter_idより大きいprogramsの1ページについて、
-- anime_idをwork_idの指す作品のanime_idと照合し、違う行だけを書き換える。work_idがNULLの行と、
-- 作品のanime_idがまだ無い行はNULLにする。updated_atは編集者が行を変えた時刻のため変えない。
-- last_idはページの最後の行のid (ページが空なら0) で、次のページのafter_idに使う。
WITH page AS (
    SELECT programs.id, works.anime_id
    FROM programs
    LEFT JOIN works ON works.id = programs.work_id
    WHERE programs.id > sqlc.arg('after_id')
    ORDER BY programs.id
    LIMIT sqlc.arg('batch_size')
), updated AS (
    UPDATE programs
    SET anime_id = page.anime_id
    FROM page
    WHERE programs.id = page.id
        AND programs.anime_id IS DISTINCT FROM page.anime_id
    RETURNING programs.id
)
SELECT
    COALESCE((SELECT MAX(page.id) FROM page), 0)::bigint AS last_id,
    (SELECT COUNT(*) FROM updated)::bigint AS updated_count;
