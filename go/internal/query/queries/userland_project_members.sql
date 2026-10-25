-- name: DeleteUserlandProjectMembersByUserID :one
-- ユーザーのUserlandプロジェクトへの参加を最大batch_size件削除し、削除した件数を返す。
WITH deleted AS (
    DELETE FROM userland_project_members
    WHERE id IN (
        SELECT target.id
        FROM userland_project_members AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING userland_project_members.id
)
SELECT COUNT(*) FROM deleted;
