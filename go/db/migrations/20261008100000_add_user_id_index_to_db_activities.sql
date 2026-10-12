-- migrate:up transaction:false
-- 退会の削除処理で、ユーザーが行ったAnnict DBの編集履歴をユーザーから引くために使う。
CREATE INDEX CONCURRENTLY IF NOT EXISTS index_db_activities_on_user_id
  ON db_activities (user_id);

-- migrate:down transaction:false
DROP INDEX CONCURRENTLY IF EXISTS index_db_activities_on_user_id;
