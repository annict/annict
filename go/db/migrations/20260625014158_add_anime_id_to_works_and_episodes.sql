-- migrate:up

-- works.anime_id / episodes.anime_idは旧work / episode行を新規採番された
-- animes行へ対応づけるマッピングカラム。animesは全行新規採番で (id空間は旧
-- works / episodesのidと範囲で分離される) ため、新旧の対応はidを再利用できず
-- 明示的なマッピングカラムで持つしかない。移行期間中はworks / episodesが書き込みの
-- 正本であり続けRailsはanime_idを設定しないため、カラムはNULL許容とし、フェーズ
-- 2のリコンシリエーション同期がanimeを作成してidを書き戻すまで行はNULLのまま。
-- 1:1のマッピングを強制するユニークインデックスは、これらの大テーブルに対して
-- ロックを保持せずCONCURRENTLYで構築するため、対になる
-- add_anime_id_index_to_works / add_anime_id_index_to_episodesマイグレーションに
-- 分けてある。
ALTER TABLE public.works ADD COLUMN anime_id BIGINT REFERENCES public.animes(id);
ALTER TABLE public.episodes ADD COLUMN anime_id BIGINT REFERENCES public.animes(id);

-- migrate:down

ALTER TABLE public.episodes DROP COLUMN IF EXISTS anime_id;
ALTER TABLE public.works DROP COLUMN IF EXISTS anime_id;
