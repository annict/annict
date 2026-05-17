-- migrate:up

-- 作品の状態を表す enum 型を作成
CREATE TYPE public.work_status AS ENUM ('published', 'archived', 'deleted');

-- works テーブルに status と archive_message カラムを追加
ALTER TABLE public.works ADD COLUMN status public.work_status NOT NULL DEFAULT 'published';
ALTER TABLE public.works ADD COLUMN archive_message VARCHAR;

-- 公開中の作品の検索用インデックス
CREATE INDEX index_works_on_status ON public.works(status) WHERE status = 'published';

-- 既存データの移行
UPDATE public.works SET status = 'deleted' WHERE deleted_at IS NOT NULL;
UPDATE public.works SET status = 'archived' WHERE unpublished_at IS NOT NULL AND deleted_at IS NULL;

-- migrate:down

-- インデックスの削除
DROP INDEX IF EXISTS public.index_works_on_status;

-- カラムの削除
ALTER TABLE public.works DROP COLUMN IF EXISTS archive_message;
ALTER TABLE public.works DROP COLUMN IF EXISTS status;

-- enum 型の削除
DROP TYPE IF EXISTS public.work_status;
