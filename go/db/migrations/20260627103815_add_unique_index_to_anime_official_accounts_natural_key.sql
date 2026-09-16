-- migrate:up

-- index_anime_official_accounts_on_anime_id_and_serviceは、フェーズ2の別表
-- リコンサイラ (SyncAnimeOfficialAccountsUsecase) が突合する (anime_id, service) の
-- 自然キーを強制する。兄弟別表 (anime_external_ids / anime_links / anime_hashtags /
-- anime_seasons) がいずれもリコンサイルのキーをUNIQUEインデックスで担保しているのに
-- 揃える。元のindex_anime_official_accounts_on_anime_id_and_serviceは非ユニークで、
-- リコンサイルが前提とするキー (worksはサービスごとに1アカウントの供給元となるため、
-- animeはサービスごとに高々1行を持つ) をDBが強制していなかった。本UNIQUE
-- インデックスは元が担っていた (anime_id, service) のルックアップをそのまま担うため、
-- 冗長な重複として残さず同じマイグレーションで旧インデックスを削除する。
--
-- anime_official_accountsはGoの同期のみが書き込む新規テーブルのため、ここでは素の
-- (トランザクション内の) CREATE INDEXで問題ない。works.anime_idのインデックスと異なり、
-- Railsと共有する大きなホットテーブルではなくCONCURRENTLYを要しない。
DROP INDEX IF EXISTS public.index_anime_official_accounts_on_anime_id_and_service;
CREATE UNIQUE INDEX index_anime_official_accounts_on_anime_id_and_service ON public.anime_official_accounts(anime_id, service);

-- migrate:down

DROP INDEX IF EXISTS public.index_anime_official_accounts_on_anime_id_and_service;
CREATE INDEX index_anime_official_accounts_on_anime_id_and_service ON public.anime_official_accounts(anime_id, service);
