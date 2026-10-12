-- name: DeletePersonFavoritesByUserID :one
-- ユーザーの人物のお気に入りを最大batch_size件削除し、削除した件数を返す。
-- お気に入り先のpeople.favorite_users_countを削除した件数だけ減算する。updated_atは変えない。
-- ユーザー自身のお気に入り数はユーザー行ごと削除するため減算しない。
WITH deleted AS (
    DELETE FROM person_favorites
    WHERE id IN (
        SELECT target.id
        FROM person_favorites AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING person_favorites.person_id
),
counts AS (
    SELECT deleted.person_id, COUNT(*) AS favorites_count
    FROM deleted
    GROUP BY deleted.person_id
),
decremented AS (
    UPDATE people
    SET favorite_users_count = people.favorite_users_count - counts.favorites_count
    FROM counts
    WHERE people.id = counts.person_id
    RETURNING people.id
)
SELECT COUNT(*) FROM deleted;
