-- migrate:up

-- Drop the partial index that covered the dormant status column.
--
-- [Ja] 休眠していた status カラムを対象にしていた部分インデックスを削除する。
DROP INDEX IF EXISTS public.index_works_on_status;

-- Drop the dormant status / archive_message columns. works state is now sourced
-- solely from unpublished_at / deleted_at (aligned with the other Unpublishable
-- resources), so these columns are no longer read by any code path.
--
-- [Ja] 休眠していた status / archive_message カラムを削除する。works の状態は
-- unpublished_at / deleted_at のみを正本とする形 (他の Unpublishable リソースと
-- 揃えた形) になり、これらのカラムはどのコード経路からも読まれなくなった。
ALTER TABLE public.works DROP COLUMN IF EXISTS archive_message;
ALTER TABLE public.works DROP COLUMN IF EXISTS status;

-- Drop the now-unused enum type. episodes keep their own episode_status type, so
-- removing work_status affects only works.
--
-- [Ja] 未使用になった enum 型を削除する。episodes は別型 episode_status を使うため、
-- work_status の削除は works のみに影響する。
DROP TYPE IF EXISTS public.work_status;

-- migrate:down

-- Recreate the enum type.
--
-- [Ja] enum 型を再作成する。
CREATE TYPE public.work_status AS ENUM ('published', 'archived', 'deleted');

-- Re-add the status / archive_message columns.
--
-- [Ja] status / archive_message カラムを再追加する。
ALTER TABLE public.works ADD COLUMN status public.work_status NOT NULL DEFAULT 'published';
ALTER TABLE public.works ADD COLUMN archive_message VARCHAR;

-- Recreate the partial index for published works.
--
-- [Ja] 公開中の作品用の部分インデックスを再作成する。
CREATE INDEX index_works_on_status ON public.works(status) WHERE status = 'published';

-- Backfill status from the timestamp columns so the restored column matches the
-- work's actual state.
--
-- [Ja] 復元した status カラムが作品の実際の状態と一致するよう、タイムスタンプ列から
-- status を復元する。
UPDATE public.works SET status = 'deleted' WHERE deleted_at IS NOT NULL;
UPDATE public.works SET status = 'archived' WHERE unpublished_at IS NOT NULL AND deleted_at IS NULL;
