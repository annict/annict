-- migrate:up transaction:false
-- REST API V2の作品の一覧を季節の順に並べるときに、代表の季節のキーの順に先頭から読むために使う。
-- キーは年 * 10 + 季節 (冬1・春2・夏3・秋4・不明5) で、一覧のクエリと同じ式にする。
CREATE INDEX CONCURRENTLY IF NOT EXISTS index_anime_seasons_on_primary_season_key_and_anime_id
  ON anime_seasons (
    (year * 10 + CASE name
      WHEN 'winter' THEN 1
      WHEN 'spring' THEN 2
      WHEN 'summer' THEN 3
      WHEN 'fall' THEN 4
      ELSE 5
    END),
    anime_id
  )
  WHERE is_primary;

-- migrate:down transaction:false
DROP INDEX CONCURRENTLY IF EXISTS index_anime_seasons_on_primary_season_key_and_anime_id;
