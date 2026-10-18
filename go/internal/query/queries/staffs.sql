-- name: GetStaffsByWorkIDs :many
-- スタッフ自身が公開状態 (unpublished_atとdeleted_atがともにNULL) の行だけを返す。
-- 人物を参照する行 (resource_type = 'Person') は、人物が閲覧できる状態 (published / archived) のものに限り、
-- 掲載停止 (delisted) と削除 (deleted) の人物の行は除外する
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
            AND p.status IN ('published', 'archived')
      )
  )
ORDER BY s.work_id, s.sort_number;
