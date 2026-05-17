-- name: GetStaffsByWorkIDs :many
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
ORDER BY s.work_id, s.sort_number;
