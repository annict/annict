-- migrate:up transaction:false
-- 退会の削除処理で、ユーザーが許可したOAuthの認可をユーザーから引くために使う。
CREATE INDEX CONCURRENTLY IF NOT EXISTS index_oauth_access_grants_on_resource_owner_id
  ON oauth_access_grants (resource_owner_id);

-- migrate:down transaction:false
DROP INDEX CONCURRENTLY IF EXISTS index_oauth_access_grants_on_resource_owner_id;
