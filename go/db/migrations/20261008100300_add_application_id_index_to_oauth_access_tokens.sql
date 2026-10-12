-- migrate:up transaction:false
-- 退会の削除処理で、論理削除済みのOAuthアプリケーションに残るアクセストークンをアプリケーションから引くために使う。
CREATE INDEX CONCURRENTLY IF NOT EXISTS index_oauth_access_tokens_on_application_id
  ON oauth_access_tokens (application_id);

-- migrate:down transaction:false
DROP INDEX CONCURRENTLY IF EXISTS index_oauth_access_tokens_on_application_id;
