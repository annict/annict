-- migrate:up transaction:false
-- series_worksの行をanimeから引くために使う。Rails版の書き込みを止めないよう、非トランザクションの
-- マイグレーションでCONCURRENTLYに作る (CONCURRENTLYは他の文と同じトランザクションに置けない)。
-- IF NOT EXISTSは、構築が中断して無効なインデックスが残ったときの再実行を冪等にする。
CREATE INDEX CONCURRENTLY IF NOT EXISTS index_series_works_on_anime_id ON public.series_works USING btree (anime_id);

-- migrate:down transaction:false
DROP INDEX CONCURRENTLY IF EXISTS public.index_series_works_on_anime_id;
