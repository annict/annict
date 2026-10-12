-- name: DeleteMuteUsersByUserID :one
-- ユーザーがしたミュートを最大batch_size件削除し、削除した件数を返す。
WITH deleted AS (
    DELETE FROM mute_users
    WHERE id IN (
        SELECT target.id
        FROM mute_users AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING mute_users.id
)
SELECT COUNT(*) FROM deleted;

-- name: DeleteMuteUsersByMutedUserID :one
-- 他のユーザーがユーザーをミュートした行を最大batch_size件削除し、削除した件数を返す。
WITH deleted AS (
    DELETE FROM mute_users
    WHERE id IN (
        SELECT target.id
        FROM mute_users AS target
        WHERE target.muted_user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING mute_users.id
)
SELECT COUNT(*) FROM deleted;
