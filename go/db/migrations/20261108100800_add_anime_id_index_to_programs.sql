-- migrate:up transaction:false
-- programsの行をanimeから引くために使う。Rails版の書き込みを止めないよう、非トランザクションの
-- マイグレーションでCONCURRENTLYに作る (CONCURRENTLYは他の文と同じトランザクションに置けない)。
-- IF NOT EXISTSは、構築が中断して無効なインデックスが残ったときの再実行を冪等にする。
CREATE INDEX CONCURRENTLY IF NOT EXISTS index_programs_on_anime_id ON public.programs USING btree (anime_id);

-- migrate:down transaction:false
DROP INDEX CONCURRENTLY IF EXISTS public.index_programs_on_anime_id;
