-- migrate:up

-- index_anime_events_on_anime_id_and_kind enforces the (anime_id, kind) natural
-- key the phase 2 satellite reconciler (SyncAnimeEventsUsecase) diffs on, matching
-- how every sibling satellite table (anime_external_ids / anime_links /
-- anime_official_accounts / anime_hashtags / anime_seasons) backs its reconcile key
-- with a UNIQUE index. The original index_anime_events_on_anime_id_and_kind was
-- non-unique, so the database did not enforce the key the reconcile assumes (works
-- source a single broadcast event, so an anime holds at most one row per kind); this
-- UNIQUE index covers the same (anime_id, kind) lookups it served, so the old index
-- is dropped in the same migration rather than kept as a redundant duplicate.
--
-- This is a migration-period constraint: works source one broadcast per anime, so the
-- key is unique while works are the source. A later phase (17) that stops referencing
-- works/episodes drops it so editors can add several rows of the same kind directly
-- (e.g. two revival_screening events); until then the reconcile relies on the key's
-- uniqueness. The same relaxation applies to the sibling anime_links /
-- anime_official_accounts UNIQUE constraints.
--
-- anime_events is a new table populated only by the Go sync, so a plain
-- (transactional) CREATE INDEX is fine here: unlike the works.anime_id index, it is
-- not a large hot table shared with Rails that would need CONCURRENTLY.
--
-- [Ja] index_anime_events_on_anime_id_and_kind は、フェーズ 2 の別表リコンサイラ
-- (SyncAnimeEventsUsecase) が突合する (anime_id, kind) の自然キーを強制する。兄弟別表
-- (anime_external_ids / anime_links / anime_official_accounts / anime_hashtags /
-- anime_seasons) がいずれもリコンサイルのキーを UNIQUE インデックスで担保しているのに
-- 揃える。元の index_anime_events_on_anime_id_and_kind は非ユニークで、リコンサイルが
-- 前提とするキー (works は broadcast イベントを 1 つ source するため、anime は kind ごとに
-- 高々 1 行を持つ) を DB が強制していなかった。本 UNIQUE インデックスは元が担っていた
-- (anime_id, kind) のルックアップをそのまま担うため、冗長な重複として残さず同じ
-- マイグレーションで旧インデックスを削除する。
--
-- これは移行期間限定の制約。works は anime ごとに broadcast を 1 つ source するため、works が
-- 源泉である間はキーが一意になる。works/episodes を参照しなくなる後続フェーズ (17) でこれを
-- DROP し、編集者が同一 kind の行を複数 (例: 2 つの revival_screening イベント) 直接足せるように
-- する。それまではリコンサイルがキーの一意性に依存する。同じ緩和は兄弟の anime_links /
-- anime_official_accounts の UNIQUE 制約にも当てはまる。
--
-- anime_events は Go の同期のみが書き込む新規テーブルのため、ここでは素の (トランザクション内の)
-- CREATE INDEX で問題ない。works.anime_id のインデックスと異なり、Rails と共有する大きな
-- ホットテーブルではなく CONCURRENTLY を要しない。
DROP INDEX IF EXISTS public.index_anime_events_on_anime_id_and_kind;
CREATE UNIQUE INDEX index_anime_events_on_anime_id_and_kind ON public.anime_events(anime_id, kind);

-- migrate:down

DROP INDEX IF EXISTS public.index_anime_events_on_anime_id_and_kind;
CREATE INDEX index_anime_events_on_anime_id_and_kind ON public.anime_events(anime_id, kind);
