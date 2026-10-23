-- name: DeleteOrganizationFavoritesByUserID :one
-- ユーザーの団体のお気に入りを最大batch_size件削除し、削除した件数を返す。
-- お気に入り先のorganizations.favorite_users_countを削除した件数だけ減算する。updated_atは変えない。
-- ユーザー自身のお気に入り数はユーザー行ごと削除するため減算しない。
WITH deleted AS (
    DELETE FROM organization_favorites
    WHERE id IN (
        SELECT target.id
        FROM organization_favorites AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING organization_favorites.organization_id
),
counts AS (
    SELECT deleted.organization_id, COUNT(*) AS favorites_count
    FROM deleted
    GROUP BY deleted.organization_id
),
decremented AS (
    UPDATE organizations
    SET favorite_users_count = organizations.favorite_users_count - counts.favorites_count
    FROM counts
    WHERE organizations.id = counts.organization_id
    RETURNING organizations.id
)
SELECT COUNT(*) FROM deleted;
