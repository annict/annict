-- name: DeleteOAuthAccessGrantsByUserID :one
-- ユーザーが許可した認可と、ユーザーが持つ論理削除済みのアプリケーションに残る
-- 他人の認可を最大batch_size件削除し、削除した件数を返す。
WITH deleted AS (
    DELETE FROM oauth_access_grants
    WHERE id IN (
        SELECT target.id
        FROM oauth_access_grants AS target
        WHERE target.resource_owner_id = sqlc.arg('user_id')
            OR target.application_id IN (
                SELECT oauth_applications.id
                FROM oauth_applications
                WHERE oauth_applications.owner_type = 'User'
                    AND oauth_applications.owner_id = sqlc.arg('user_id')::bigint
                    AND oauth_applications.deleted_at IS NOT NULL
            )
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING oauth_access_grants.id
)
SELECT COUNT(*) FROM deleted;
