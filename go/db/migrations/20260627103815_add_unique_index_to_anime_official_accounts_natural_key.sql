-- migrate:up

-- index_anime_official_accounts_on_anime_id_and_service enforces the
-- (anime_id, service) natural key the phase 2 satellite reconciler
-- (SyncAnimeOfficialAccountsUsecase) diffs on, matching how every sibling
-- satellite table (anime_external_ids / anime_links / anime_hashtags /
-- anime_seasons) backs its reconcile key with a UNIQUE index. The original
-- index_anime_official_accounts_on_anime_id_and_service was non-unique, so the
-- database did not enforce the key the reconcile assumes (works source a single
-- account per service, so an anime holds at most one row per service); this
-- UNIQUE index covers the same (anime_id, service) lookups it served, so the old
-- index is dropped in the same migration rather than kept as a redundant duplicate.
--
-- anime_official_accounts is a new table populated only by the Go sync, so a plain
-- (transactional) CREATE INDEX is fine here: unlike the works.anime_id index, it is
-- not a large hot table shared with Rails that would need CONCURRENTLY.
--
-- [Ja] index_anime_official_accounts_on_anime_id_and_service は、フェーズ 2 の別表
-- リコンサイラ (SyncAnimeOfficialAccountsUsecase) が突合する (anime_id, service) の
-- 自然キーを強制する。兄弟別表 (anime_external_ids / anime_links / anime_hashtags /
-- anime_seasons) がいずれもリコンサイルのキーを UNIQUE インデックスで担保しているのに
-- 揃える。元の index_anime_official_accounts_on_anime_id_and_service は非ユニークで、
-- リコンサイルが前提とするキー (works はサービスごとに 1 アカウントを source するため、
-- anime はサービスごとに高々 1 行を持つ) を DB が強制していなかった。本 UNIQUE
-- インデックスは元が担っていた (anime_id, service) のルックアップをそのまま担うため、
-- 冗長な重複として残さず同じマイグレーションで旧インデックスを削除する。
--
-- anime_official_accounts は Go の同期のみが書き込む新規テーブルのため、ここでは素の
-- (トランザクション内の) CREATE INDEX で問題ない。works.anime_id のインデックスと異なり、
-- Rails と共有する大きなホットテーブルではなく CONCURRENTLY を要しない。
DROP INDEX IF EXISTS public.index_anime_official_accounts_on_anime_id_and_service;
CREATE UNIQUE INDEX index_anime_official_accounts_on_anime_id_and_service ON public.anime_official_accounts(anime_id, service);

-- migrate:down

DROP INDEX IF EXISTS public.index_anime_official_accounts_on_anime_id_and_service;
CREATE INDEX index_anime_official_accounts_on_anime_id_and_service ON public.anime_official_accounts(anime_id, service);
