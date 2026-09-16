-- migrate:up

-- 休眠していたstatusカラムを対象にしていた部分インデックスを削除する。
DROP INDEX IF EXISTS public.index_episodes_on_status;

-- 休眠していたstatus / archive_messageカラムを削除する。エピソードの状態は
-- unpublished_at / deleted_atのみを正本とする形 (worksや他のUnpublishable
-- リソースと揃えた形) になり、これらのカラムはどのコード経路からも読まれなくなった。
ALTER TABLE public.episodes DROP COLUMN IF EXISTS archive_message;
ALTER TABLE public.episodes DROP COLUMN IF EXISTS status;

-- 未使用になったenum型を削除する。work_statusはworksのカラム削除と併せて
-- 削除済みのため、対になるepisode_statusが最後の1つとなる。
DROP TYPE IF EXISTS public.episode_status;

-- migrate:down

-- enum型を再作成する。
CREATE TYPE public.episode_status AS ENUM ('published', 'archived', 'deleted');

-- status / archive_messageカラムを再追加する。
ALTER TABLE public.episodes ADD COLUMN status public.episode_status NOT NULL DEFAULT 'published';
ALTER TABLE public.episodes ADD COLUMN archive_message VARCHAR;

-- 公開中のエピソード用の部分インデックスを再作成する。
CREATE INDEX index_episodes_on_status ON public.episodes(status) WHERE status = 'published';

-- 復元したstatusカラムがエピソードの実際の状態と一致するよう、タイムスタンプ列から
-- statusを復元する。
UPDATE public.episodes SET status = 'deleted' WHERE deleted_at IS NOT NULL;
UPDATE public.episodes SET status = 'archived' WHERE unpublished_at IS NOT NULL AND deleted_at IS NULL;
