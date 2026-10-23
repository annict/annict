-- name: DeleteFollowsByUserID :one
-- ユーザーがしたフォローを最大batch_size件削除し、削除した件数を返す。
-- フォロー先のusers.followers_countを削除した件数だけ減算する。updated_atは変えない。
-- フォローの通知は先の段階 (ユーザーが起こした通知の削除) で削除しておく。
WITH deleted AS (
    DELETE FROM follows
    WHERE id IN (
        SELECT target.id
        FROM follows AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING follows.following_id
),
counts AS (
    SELECT deleted.following_id, COUNT(*) AS follows_count
    FROM deleted
    GROUP BY deleted.following_id
),
decremented AS (
    UPDATE users
    SET followers_count = users.followers_count - counts.follows_count
    FROM counts
    WHERE users.id = counts.following_id
    RETURNING users.id
)
SELECT COUNT(*) FROM deleted;

-- name: DeleteFollowsByFollowingID :one
-- ユーザーへのフォローを最大batch_size件削除し、削除した件数を返す。
-- フォローしていたユーザーのusers.following_countを削除した件数だけ減算する。updated_atは変えない。
-- フォローの通知は先の段階 (ユーザー宛ての通知の削除) で削除しておく。
WITH deleted AS (
    DELETE FROM follows
    WHERE id IN (
        SELECT target.id
        FROM follows AS target
        WHERE target.following_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING follows.user_id
),
counts AS (
    SELECT deleted.user_id, COUNT(*) AS follows_count
    FROM deleted
    GROUP BY deleted.user_id
),
decremented AS (
    UPDATE users
    SET following_count = users.following_count - counts.follows_count
    FROM counts
    WHERE users.id = counts.user_id
    RETURNING users.id
)
SELECT COUNT(*) FROM deleted;
