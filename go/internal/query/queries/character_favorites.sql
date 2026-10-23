-- name: DeleteCharacterFavoritesByUserID :one
-- ユーザーのキャラクターのお気に入りを最大batch_size件削除し、削除した件数を返す。
-- お気に入り先のcharacters.favorite_users_countを削除した件数だけ減算する。updated_atは変えない。
-- ユーザー自身のお気に入り数はユーザー行ごと削除するため減算しない。
WITH deleted AS (
    DELETE FROM character_favorites
    WHERE id IN (
        SELECT target.id
        FROM character_favorites AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING character_favorites.character_id
),
counts AS (
    SELECT deleted.character_id, COUNT(*) AS favorites_count
    FROM deleted
    GROUP BY deleted.character_id
),
decremented AS (
    UPDATE characters
    SET favorite_users_count = characters.favorite_users_count - counts.favorites_count
    FROM counts
    WHERE characters.id = counts.character_id
    RETURNING characters.id
)
SELECT COUNT(*) FROM deleted;
