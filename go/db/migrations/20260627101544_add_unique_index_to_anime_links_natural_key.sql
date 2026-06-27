-- migrate:up

-- index_anime_links_on_anime_id_and_kind_and_language enforces the
-- (anime_id, kind, language) natural key the phase 2 satellite reconciler
-- (SyncAnimeLinksUsecase) diffs on, matching how every sibling satellite table
-- (anime_external_ids / anime_hashtags / anime_seasons) backs its reconcile key
-- with a UNIQUE index. The original index_anime_links_on_anime_id_and_kind was
-- non-unique and omitted language, so the database did not enforce the key the
-- reconcile assumes; this UNIQUE index is a strict superset of it (its
-- (anime_id, kind) prefix serves the same lookups), so the old index is dropped
-- in the same migration rather than kept as a redundant duplicate.
--
-- anime_links is a new table populated only by the Go sync, so a plain
-- (transactional) CREATE INDEX is fine here: unlike the works.anime_id index, it
-- is not a large hot table shared with Rails that would need CONCURRENTLY.
--
-- [Ja] index_anime_links_on_anime_id_and_kind_and_language は、フェーズ 2 の別表
-- リコンサイラ (SyncAnimeLinksUsecase) が突合する (anime_id, kind, language) の
-- 自然キーを強制する。兄弟別表 (anime_external_ids / anime_hashtags /
-- anime_seasons) がいずれもリコンサイルのキーを UNIQUE インデックスで担保しているのに
-- 揃える。元の index_anime_links_on_anime_id_and_kind は非ユニークで language を
-- 含まず、リコンサイルが前提とするキーを DB が強制していなかった。本 UNIQUE
-- インデックスはその厳密な上位集合 ((anime_id, kind) プレフィックスが同じルックアップを
-- 担う) のため、冗長な重複として残さず同じマイグレーションで旧インデックスを削除する。
--
-- anime_links は Go の同期のみが書き込む新規テーブルのため、ここでは素の (トランザク
-- ション内の) CREATE INDEX で問題ない。works.anime_id のインデックスと異なり、Rails と
-- 共有する大きなホットテーブルではなく CONCURRENTLY を要しない。
DROP INDEX IF EXISTS public.index_anime_links_on_anime_id_and_kind;
CREATE UNIQUE INDEX index_anime_links_on_anime_id_and_kind_and_language ON public.anime_links(anime_id, kind, language);

-- migrate:down

DROP INDEX IF EXISTS public.index_anime_links_on_anime_id_and_kind_and_language;
CREATE INDEX index_anime_links_on_anime_id_and_kind ON public.anime_links(anime_id, kind);
