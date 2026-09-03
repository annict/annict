-- migrate:up transaction:false

-- index_works_on_anime_id enforces the one-to-one legacy <-> animes mapping
-- carried by works.anime_id (added in the companion
-- add_anime_id_to_works_and_episodes migration). It is partial
-- (WHERE anime_id IS NOT NULL) so the many not-yet-synced NULL rows are not
-- indexed while a one-to-one mapping is still enforced on the synced rows.
--
-- works is a large existing table shared with Rails, so the index is built
-- CONCURRENTLY in a non-transactional migration: a plain CREATE INDEX would hold
-- a lock on this hot catalog table for the duration of a full heap scan. The
-- episodes index lives in its own migration because CREATE INDEX CONCURRENTLY
-- cannot share a migration with another statement (multiple statements run in an
-- implicit transaction, which CONCURRENTLY rejects). IF NOT EXISTS / IF EXISTS
-- keep re-runs idempotent if a CONCURRENTLY build is interrupted and leaves an
-- invalid index behind.
--
-- [Ja] index_works_on_anime_id は、works.anime_id (companion の
-- add_anime_id_to_works_and_episodes マイグレーションで追加) が持つ旧 <-> animes
-- の 1:1 マッピングを強制する。部分インデックス (WHERE anime_id IS NOT NULL) とし、
-- 未同期で NULL の多数の行をインデックスせずに、同期済み行の 1:1 マッピングだけを
-- 強制する。
--
-- works は Rails と共有する既存の大テーブルのため、インデックスは非トランザクションの
-- マイグレーションで CONCURRENTLY 構築する。素の CREATE INDEX だと、フルヒープ
-- スキャンの間この参照の多いカタログテーブルにロックを保持してしまう。episodes の
-- インデックスを別マイグレーションに分けているのは、CREATE INDEX CONCURRENTLY が
-- 他の文と同じマイグレーションに同居できないため (複数文は暗黙トランザクションで実行
-- され、CONCURRENTLY がそれを拒否する)。IF NOT EXISTS / IF EXISTS は、CONCURRENTLY
-- の構築が中断して無効なインデックスを残したときの再実行を冪等にする。
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS index_works_on_anime_id ON public.works(anime_id) WHERE anime_id IS NOT NULL;

-- migrate:down transaction:false

DROP INDEX CONCURRENTLY IF EXISTS public.index_works_on_anime_id;
