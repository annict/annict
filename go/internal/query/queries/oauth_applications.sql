-- name: DeleteDeletedOAuthApplicationsByOwnerID :one
-- ユーザーが持つ論理削除済みのアプリケーションを最大batch_size件削除し、削除した件数を返す。
-- アプリケーションを参照するアクセストークン・認可・記録・ステータスは先に削除するか、
-- 参照を外しておく (外部キーがあるため)。論理削除されていないアプリケーションは扱わない。
WITH deleted AS (
    DELETE FROM oauth_applications
    WHERE id IN (
        SELECT target.id
        FROM oauth_applications AS target
        WHERE target.owner_type = 'User'
            AND target.owner_id = sqlc.arg('user_id')::bigint
            AND target.deleted_at IS NOT NULL
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING oauth_applications.id
)
SELECT COUNT(*) FROM deleted;

-- name: ExistsKeptOAuthApplicationByOwnerID :one
-- ユーザーが持つ論理削除されていないアプリケーションがあるかを返す。
SELECT EXISTS (
    SELECT 1
    FROM oauth_applications
    WHERE owner_type = 'User'
        AND owner_id = sqlc.arg('user_id')::bigint
        AND deleted_at IS NULL
);
