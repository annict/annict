-- name: GetCastsByWorkIDs :many
-- キャスト自身が公開状態 (unpublished_atとdeleted_atがともにNULL) で、人物が閲覧できる状態
-- (published / archived) の行だけを返す。掲載停止 (delisted) と削除 (deleted) の人物の行は除外する
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
  AND p.status IN ('published', 'archived')
ORDER BY c.work_id, c.sort_number;
