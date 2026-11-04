-- name: DeleteOAuthAccessTokensByUserID :one
-- ユーザーに発行したアクセストークンと、ユーザーが持つ論理削除済みのアプリケーションに残る
-- 他人のアクセストークンを最大batch_size件削除し、削除した件数を返す。
WITH deleted AS (
    DELETE FROM oauth_access_tokens
    WHERE id IN (
        SELECT target.id
        FROM oauth_access_tokens AS target
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
    RETURNING oauth_access_tokens.id
)
SELECT COUNT(*) FROM deleted;

-- name: GetOAuthAccessTokenByToken :one
-- REST API V2で受け取ったアクセストークンを照合するために引く。
-- 退会・利用停止でdeleted_atが立った利用者のトークンは返さない。
SELECT
    oauth_access_tokens.id,
    oauth_access_tokens.resource_owner_id,
    oauth_access_tokens.application_id,
    oauth_access_tokens.expires_in,
    oauth_access_tokens.revoked_at,
    oauth_access_tokens.created_at,
    oauth_access_tokens.scopes
FROM oauth_access_tokens
INNER JOIN users ON users.id = oauth_access_tokens.resource_owner_id
WHERE oauth_access_tokens.token = sqlc.arg('token')
    AND users.deleted_at IS NULL;
