-- migrate:up transaction:false

-- index_works_on_anime_idは、works.anime_id (対になる
-- add_anime_id_to_works_and_episodesマイグレーションで追加) が持つ旧 <-> animes
-- の1:1マッピングを強制する。部分インデックス (WHERE anime_id IS NOT NULL) とし、
-- 未同期でNULLの多数の行をインデックスせずに、同期済み行の1:1マッピングだけを
-- 強制する。
--
-- worksはRailsと共有する既存の大テーブルのため、インデックスは非トランザクションの
-- マイグレーションでCONCURRENTLY構築する。素のCREATE INDEXだと、フルヒープ
-- スキャンの間この参照の多いカタログテーブルにロックを保持してしまう。episodesの
-- インデックスを別マイグレーションに分けているのは、CREATE INDEX CONCURRENTLYが
-- 他の文と同じマイグレーションに同居できないため (複数文は暗黙トランザクションで実行
-- され、CONCURRENTLYがそれを拒否する)。IF NOT EXISTS / IF EXISTSは、CONCURRENTLY
-- の構築が中断して無効なインデックスを残したときの再実行を冪等にする。
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS index_works_on_anime_id ON public.works(anime_id) WHERE anime_id IS NOT NULL;

-- migrate:down transaction:false

DROP INDEX CONCURRENTLY IF EXISTS public.index_works_on_anime_id;
