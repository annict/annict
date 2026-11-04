-- migrate:up transaction:false
-- REST API V2の作品の一覧で、エピソードを含むanimeの全体ではなく作品だけを読むために使う。
CREATE INDEX CONCURRENTLY IF NOT EXISTS index_anime_classifications_on_anime_id_of_works
  ON anime_classifications (anime_id)
  WHERE kind = 'work';

-- migrate:down transaction:false
DROP INDEX CONCURRENTLY IF EXISTS index_anime_classifications_on_anime_id_of_works;
