-- migrate:up

-- works.anime_id / episodes.anime_id map each legacy work / episode row to its
-- newly assigned animes row. animes are all freshly numbered (their id space is
-- range-separated from the legacy works / episodes ids), so the old <-> new
-- correspondence cannot reuse the id and has to live in an explicit mapping
-- column. The column is nullable because works / episodes stay the write source
-- of truth during the migration and Rails never sets anime_id; a row stays NULL
-- until the phase 2 reconciliation sync creates its anime and writes the id
-- back. The unique indexes that enforce the one-to-one mapping live in the
-- companion add_anime_id_index_to_works and add_anime_id_index_to_episodes
-- migrations so they can be built CONCURRENTLY against these large tables
-- without holding a lock.
--
-- [Ja] works.anime_id / episodes.anime_id は旧 work / episode 行を新規採番された
-- animes 行へ対応づけるマッピングカラム。animes は全行新規採番で (id 空間は旧
-- works / episodes の id と範囲で分離される) ため、新旧の対応は id を再利用できず
-- 明示的なマッピングカラムで持つしかない。移行期間中は works / episodes が書き込みの
-- 正本であり続け Rails は anime_id を設定しないため、カラムは NULL 許容とし、フェーズ
-- 2 のリコンシリエーション同期が anime を作成して id を書き戻すまで行は NULL のまま。
-- 1:1 のマッピングを強制するユニークインデックスは、これらの大テーブルに対して
-- ロックを保持せず CONCURRENTLY で構築するため、companion の
-- add_anime_id_index_to_works / add_anime_id_index_to_episodes マイグレーションに
-- 分けてある。
ALTER TABLE public.works ADD COLUMN anime_id BIGINT REFERENCES public.animes(id);
ALTER TABLE public.episodes ADD COLUMN anime_id BIGINT REFERENCES public.animes(id);

-- migrate:down

ALTER TABLE public.episodes DROP COLUMN IF EXISTS anime_id;
ALTER TABLE public.works DROP COLUMN IF EXISTS anime_id;
