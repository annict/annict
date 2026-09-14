-- migrate:up

-- 休眠していたstatusカラムを対象にしていた部分インデックスを削除する。
DROP INDEX IF EXISTS public.index_works_on_status;

-- 休眠していたstatus / archive_messageカラムを削除する。worksの状態は
-- unpublished_at / deleted_atのみを正本とする形 (他のUnpublishableリソースと
-- 揃えた形) になり、これらのカラムはどのコード経路からも読まれなくなった。
ALTER TABLE public.works DROP COLUMN IF EXISTS archive_message;
ALTER TABLE public.works DROP COLUMN IF EXISTS status;

-- 未使用になったenum型を削除する。episodesは別型episode_statusを使うため、
-- work_statusの削除はworksのみに影響する。
DROP TYPE IF EXISTS public.work_status;

-- migrate:down

-- enum型を再作成する。
CREATE TYPE public.work_status AS ENUM ('published', 'archived', 'deleted');

-- status / archive_messageカラムを再追加する。
ALTER TABLE public.works ADD COLUMN status public.work_status NOT NULL DEFAULT 'published';
ALTER TABLE public.works ADD COLUMN archive_message VARCHAR;

-- 公開中の作品用の部分インデックスを再作成する。
CREATE INDEX index_works_on_status ON public.works(status) WHERE status = 'published';

-- 復元したstatusカラムが作品の実際の状態と一致するよう、タイムスタンプ列から
-- statusを復元する。
UPDATE public.works SET status = 'deleted' WHERE deleted_at IS NOT NULL;
UPDATE public.works SET status = 'archived' WHERE unpublished_at IS NOT NULL AND deleted_at IS NULL;
