-- migrate:up

-- index_anime_events_on_anime_id_and_kindは、フェーズ2の別表リコンサイラ
-- (SyncAnimeEventsUsecase) が突合する (anime_id, kind) の自然キーを強制する。兄弟別表
-- (anime_external_ids / anime_links / anime_official_accounts / anime_hashtags /
-- anime_seasons) がいずれもリコンサイルのキーをUNIQUEインデックスで担保しているのに
-- 揃える。元のindex_anime_events_on_anime_id_and_kindは非ユニークで、リコンサイルが
-- 前提とするキー (worksはbroadcastイベント1つの供給元となるため、animeはkindごとに
-- 高々1行を持つ) をDBが強制していなかった。本UNIQUEインデックスは元が担っていた
-- (anime_id, kind) のルックアップをそのまま担うため、冗長な重複として残さず同じ
-- マイグレーションで旧インデックスを削除する。
--
-- これは移行期間限定の制約。worksはanimeごとにbroadcastイベント1つの供給元となるため、worksが
-- 源泉である間はキーが一意になる。works/episodesを参照しなくなる後続フェーズ (17) でこれを
-- DROPし、編集者が同一kindの行を複数 (例: 2つのrevival_screeningイベント) 直接足せるように
-- する。それまではリコンサイルがキーの一意性に依存する。同じ緩和は兄弟のanime_links /
-- anime_official_accountsのUNIQUE制約にも当てはまる。
--
-- anime_eventsはGoの同期のみが書き込む新規テーブルのため、ここでは素の (トランザクション内の)
-- CREATE INDEXで問題ない。works.anime_idのインデックスと異なり、Railsと共有する大きな
-- ホットテーブルではなくCONCURRENTLYを要しない。
DROP INDEX IF EXISTS public.index_anime_events_on_anime_id_and_kind;
CREATE UNIQUE INDEX index_anime_events_on_anime_id_and_kind ON public.anime_events(anime_id, kind);

-- migrate:down

DROP INDEX IF EXISTS public.index_anime_events_on_anime_id_and_kind;
CREATE INDEX index_anime_events_on_anime_id_and_kind ON public.anime_events(anime_id, kind);
