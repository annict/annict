-- migrate:up transaction:false
-- 退会の削除処理で、ステータスや記録に付いたいいねをいいね先から引くために使う。
CREATE INDEX CONCURRENTLY IF NOT EXISTS index_likes_on_recipient_type_and_recipient_id
  ON likes (recipient_type, recipient_id);

-- migrate:down transaction:false
DROP INDEX CONCURRENTLY IF EXISTS index_likes_on_recipient_type_and_recipient_id;
