-- name: DeleteForumCommentsOnForumPostsByUserID :one
-- ユーザーが立てたスレッドに付いたコメント (他人のものを含む) を最大batch_size件削除し、
-- 削除した件数を返す。スレッドごと削除するため、スレッドのコメント数は減算しない。
WITH deleted AS (
    DELETE FROM forum_comments
    WHERE id IN (
        SELECT target.id
        FROM forum_comments AS target
        WHERE target.forum_post_id IN (
            SELECT forum_posts.id
            FROM forum_posts
            WHERE forum_posts.user_id = sqlc.arg('user_id')
        )
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING forum_comments.id
)
SELECT COUNT(*) FROM deleted;

-- name: DetachForumCommentsFromUser :one
-- ユーザーが他人のスレッドに書いたコメントを最大batch_size件、投稿者との紐づきを外し、
-- 外した件数を返す。スレッドの流れを保つため本文は残す。ユーザーが立てたスレッドの
-- コメントは先に削除しておく。
WITH detached AS (
    UPDATE forum_comments
    SET user_id = NULL
    WHERE id IN (
        SELECT target.id
        FROM forum_comments AS target
        WHERE target.user_id = sqlc.arg('user_id')::bigint
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING forum_comments.id
)
SELECT COUNT(*) FROM detached;
