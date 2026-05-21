-- migrate:up transaction:false
CREATE INDEX CONCURRENTLY IF NOT EXISTS index_records_on_user_id_and_watched_at
  ON records (user_id, watched_at)
  WHERE deleted_at IS NULL;

-- migrate:down transaction:false
DROP INDEX CONCURRENTLY IF EXISTS index_records_on_user_id_and_watched_at;
