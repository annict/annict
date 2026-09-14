-- migrate:up transaction:false

-- index_episodes_on_anime_idはindex_works_on_anime_idのepisodes版。
-- 部分インデックスとCONCURRENTLYの理由はadd_anime_id_index_to_works
-- マイグレーションを参照。CREATE INDEX CONCURRENTLYは他の文と同じマイグレーションに
-- 同居できないため、独立したマイグレーションにしている。
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS index_episodes_on_anime_id ON public.episodes(anime_id) WHERE anime_id IS NOT NULL;

-- migrate:down transaction:false

DROP INDEX CONCURRENTLY IF EXISTS public.index_episodes_on_anime_id;
