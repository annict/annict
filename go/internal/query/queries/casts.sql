-- name: GetCastsByWorkIDs :many
SELECT
    c.id,
    c.work_id,
    c.name,
    c.name_en,
    c.sort_number,
    ch.name as character_name,
    ch.name_en as character_name_en,
    p.name as person_name,
    p.name_en as person_name_en
FROM casts c
LEFT JOIN characters ch ON c.character_id = ch.id
LEFT JOIN people p ON c.person_id = p.id
WHERE c.work_id = ANY($1::bigint[])
ORDER BY c.work_id, c.sort_number;
