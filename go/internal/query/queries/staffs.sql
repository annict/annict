-- name: GetStaffsByWorkIDs :many
-- スタッフ自身が公開状態 (unpublished_atとdeleted_atがともにNULL) の行だけを返す。
-- 人物を参照する行 (resource_type = 'Person') は、人物が公開中 (published) のものに限り、
-- 非公開 (archived)・掲載停止 (delisted)・削除 (deleted) の人物の行は、Rails版の作品ページと同じく除外する
SELECT
    s.id,
    s.work_id,
    s.name,
    s.name_en,
    s.role,
    s.role_other,
    s.role_other_en,
    s.sort_number
FROM staffs s
WHERE s.work_id = ANY($1::bigint[])
  AND s.role != 'other'
  AND s.deleted_at IS NULL
  AND s.unpublished_at IS NULL
  AND (
      s.resource_type != 'Person'
      OR EXISTS (
          SELECT 1
          FROM people p
          WHERE p.id = s.resource_id
            AND p.status = 'published'
      )
  )
ORDER BY s.work_id, s.sort_number;

-- name: ReconcileStaffAnimeIDsAfter :one
-- animes同期バッチの関連テーブルの照合。idがafter_idより大きいstaffsの1ページについて、
-- anime_idをwork_idの指す作品のanime_idと照合し、違う行だけを書き換える。work_idがNULLの行と、
-- 作品のanime_idがまだ無い行はNULLにする。updated_atは編集者が行を変えた時刻のため変えない。
-- last_idはページの最後の行のid (ページが空なら0) で、次のページのafter_idに使う。
WITH page AS (
    SELECT staffs.id, works.anime_id
    FROM staffs
    LEFT JOIN works ON works.id = staffs.work_id
    WHERE staffs.id > sqlc.arg('after_id')
    ORDER BY staffs.id
    LIMIT sqlc.arg('batch_size')
), updated AS (
    UPDATE staffs
    SET anime_id = page.anime_id
    FROM page
    WHERE staffs.id = page.id
        AND staffs.anime_id IS DISTINCT FROM page.anime_id
    RETURNING staffs.id
)
SELECT
    COALESCE((SELECT MAX(page.id) FROM page), 0)::bigint AS last_id,
    (SELECT COUNT(*) FROM updated)::bigint AS updated_count;
