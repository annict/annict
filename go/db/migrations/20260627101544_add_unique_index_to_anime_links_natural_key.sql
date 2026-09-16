-- migrate:up

-- index_anime_links_on_anime_id_and_kind_and_languageは、フェーズ2の別表
-- リコンサイラ (SyncAnimeLinksUsecase) が突合する (anime_id, kind, language) の
-- 自然キーを強制する。兄弟別表 (anime_external_ids / anime_hashtags /
-- anime_seasons) がいずれもリコンサイルのキーをUNIQUEインデックスで担保しているのに
-- 揃える。元のindex_anime_links_on_anime_id_and_kindは非ユニークでlanguageを
-- 含まず、リコンサイルが前提とするキーをDBが強制していなかった。本UNIQUE
-- インデックスはその厳密な上位集合 ((anime_id, kind) プレフィックスが同じルックアップを
-- 担う) のため、冗長な重複として残さず同じマイグレーションで旧インデックスを削除する。
--
-- anime_linksはGoの同期のみが書き込む新規テーブルのため、ここでは素の (トランザク
-- ション内の) CREATE INDEXで問題ない。works.anime_idのインデックスと異なり、Railsと
-- 共有する大きなホットテーブルではなくCONCURRENTLYを要しない。
DROP INDEX IF EXISTS public.index_anime_links_on_anime_id_and_kind;
CREATE UNIQUE INDEX index_anime_links_on_anime_id_and_kind_and_language ON public.anime_links(anime_id, kind, language);

-- migrate:down

DROP INDEX IF EXISTS public.index_anime_links_on_anime_id_and_kind_and_language;
CREATE INDEX index_anime_links_on_anime_id_and_kind ON public.anime_links(anime_id, kind);
