-- name: ListPrefectures :many
-- 人物の編集フォームの出身地の選択肢。都道府県の順 (北海道から沖縄県、最後に海外) は
-- IDの順と一致するため、IDで並べる。
SELECT id, name
FROM prefectures
ORDER BY id;

-- name: ExistsPrefectureByID :one
SELECT EXISTS (
    SELECT 1
    FROM prefectures
    WHERE id = sqlc.arg('id')
);
