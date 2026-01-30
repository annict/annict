-- migrate:up

-- works テーブルに anime_id カラムを追加
ALTER TABLE works ADD COLUMN anime_id BIGINT REFERENCES animes(id);
CREATE UNIQUE INDEX index_works_on_anime_id ON works(anime_id) WHERE anime_id IS NOT NULL;

-- episodes テーブルに anime_id カラムを追加
ALTER TABLE episodes ADD COLUMN anime_id BIGINT REFERENCES animes(id);
CREATE UNIQUE INDEX index_episodes_on_anime_id ON episodes(anime_id) WHERE anime_id IS NOT NULL;

-- episodes テーブルに prev_anime_id カラムを追加
ALTER TABLE episodes ADD COLUMN prev_anime_id BIGINT REFERENCES animes(id);
CREATE UNIQUE INDEX index_episodes_on_prev_anime_id ON episodes(prev_anime_id) WHERE prev_anime_id IS NOT NULL;

-- migrate:down

DROP INDEX IF EXISTS index_episodes_on_prev_anime_id;
ALTER TABLE episodes DROP COLUMN IF EXISTS prev_anime_id;

DROP INDEX IF EXISTS index_episodes_on_anime_id;
ALTER TABLE episodes DROP COLUMN IF EXISTS anime_id;

DROP INDEX IF EXISTS index_works_on_anime_id;
ALTER TABLE works DROP COLUMN IF EXISTS anime_id;
