-- name: DeleteWorkTaggingsByUserID :one
-- ユーザーが作品に付けたタグを最大batch_size件削除し、削除した件数を返す。
-- タグのwork_tags.work_taggings_countを削除した件数だけ減算する。updated_atは変えない。
WITH deleted AS (
    DELETE FROM work_taggings
    WHERE id IN (
        SELECT target.id
        FROM work_taggings AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING work_taggings.work_tag_id
),
counts AS (
    SELECT deleted.work_tag_id, COUNT(*) AS work_taggings_count
    FROM deleted
    GROUP BY deleted.work_tag_id
),
decremented AS (
    UPDATE work_tags
    SET work_taggings_count = work_tags.work_taggings_count - counts.work_taggings_count
    FROM counts
    WHERE work_tags.id = counts.work_tag_id
    RETURNING work_tags.id
)
SELECT COUNT(*) FROM deleted;
