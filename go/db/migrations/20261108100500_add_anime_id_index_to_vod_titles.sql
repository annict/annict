-- migrate:up transaction:false
-- vod_titlesの行をanimeから引くために使う。Rails版の書き込みを止めないよう、非トランザクションの
-- マイグレーションでCONCURRENTLYに作る (CONCURRENTLYは他の文と同じトランザクションに置けない)。
-- IF NOT EXISTSは作成済みのインデックスがある場合の再実行を許す。
-- 中断で同名の無効なインデックスが残った場合は再構築されないため、無効であることを
-- 確認し、DROP INDEX CONCURRENTLYで削除してから再実行する。
CREATE INDEX CONCURRENTLY IF NOT EXISTS index_vod_titles_on_anime_id ON public.vod_titles USING btree (anime_id);

-- migrate:down transaction:false
DROP INDEX CONCURRENTLY IF EXISTS public.index_vod_titles_on_anime_id;
