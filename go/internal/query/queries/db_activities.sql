-- name: DeleteDbActivitiesByUserID :one
-- ユーザーが行ったAnnict DBの編集履歴を最大batch_size件削除し、削除した件数を返す。
WITH deleted AS (
    DELETE FROM db_activities
    WHERE id IN (
        SELECT target.id
        FROM db_activities AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING db_activities.id
)
SELECT COUNT(*) FROM deleted;
