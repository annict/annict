-- name: DeleteForumPostParticipantsByUserID :one
-- ユーザーのスレッドへの参加と、ユーザーが立てたスレッドの他人の参加を最大batch_size件削除し、
-- 削除した件数を返す。
WITH deleted AS (
    DELETE FROM forum_post_participants
    WHERE id IN (
        SELECT target.id
        FROM forum_post_participants AS target
        WHERE target.user_id = sqlc.arg('user_id')
            OR target.forum_post_id IN (
                SELECT forum_posts.id
                FROM forum_posts
                WHERE forum_posts.user_id = sqlc.arg('user_id')
            )
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING forum_post_participants.id
)
SELECT COUNT(*) FROM deleted;
