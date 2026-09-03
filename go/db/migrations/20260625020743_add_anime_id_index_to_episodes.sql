-- migrate:up transaction:false

-- index_episodes_on_anime_id is the episodes counterpart of
-- index_works_on_anime_id; see the add_anime_id_index_to_works migration for the
-- partial-index and CONCURRENTLY rationale. It lives in its own migration
-- because CREATE INDEX CONCURRENTLY cannot share a migration with another
-- statement.
--
-- [Ja] index_episodes_on_anime_id は index_works_on_anime_id の episodes 版。
-- 部分インデックスと CONCURRENTLY の理由は add_anime_id_index_to_works
-- マイグレーションを参照。CREATE INDEX CONCURRENTLY は他の文と同じマイグレーションに
-- 同居できないため、独立したマイグレーションにしている。
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS index_episodes_on_anime_id ON public.episodes(anime_id) WHERE anime_id IS NOT NULL;

-- migrate:down transaction:false

DROP INDEX CONCURRENTLY IF EXISTS public.index_episodes_on_anime_id;
