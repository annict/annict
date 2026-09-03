-- migrate:up

-- Drop the partial index that covered the dormant status column.
--
-- [Ja] 休眠していた status カラムを対象にしていた部分インデックスを削除する。
DROP INDEX IF EXISTS public.index_episodes_on_status;

-- Drop the dormant status / archive_message columns. An episode's state is now
-- sourced solely from unpublished_at / deleted_at (aligned with works and the
-- other Unpublishable resources), so these columns are no longer read by any
-- code path.
--
-- [Ja] 休眠していた status / archive_message カラムを削除する。エピソードの状態は
-- unpublished_at / deleted_at のみを正本とする形 (works や他の Unpublishable
-- リソースと揃えた形) になり、これらのカラムはどのコード経路からも読まれなくなった。
ALTER TABLE public.episodes DROP COLUMN IF EXISTS archive_message;
ALTER TABLE public.episodes DROP COLUMN IF EXISTS status;

-- Drop the now-unused enum type. work_status was already dropped alongside the
-- works columns, so episode_status is the last of the pair to go.
--
-- [Ja] 未使用になった enum 型を削除する。work_status は works のカラム削除と併せて
-- 削除済みのため、対になる episode_status が最後の 1 つとなる。
DROP TYPE IF EXISTS public.episode_status;

-- migrate:down

-- Recreate the enum type.
--
-- [Ja] enum 型を再作成する。
CREATE TYPE public.episode_status AS ENUM ('published', 'archived', 'deleted');

-- Re-add the status / archive_message columns.
--
-- [Ja] status / archive_message カラムを再追加する。
ALTER TABLE public.episodes ADD COLUMN status public.episode_status NOT NULL DEFAULT 'published';
ALTER TABLE public.episodes ADD COLUMN archive_message VARCHAR;

-- Recreate the partial index for published episodes.
--
-- [Ja] 公開中のエピソード用の部分インデックスを再作成する。
CREATE INDEX index_episodes_on_status ON public.episodes(status) WHERE status = 'published';

-- Backfill status from the timestamp columns so the restored column matches the
-- episode's actual state.
--
-- [Ja] 復元した status カラムがエピソードの実際の状態と一致するよう、タイムスタンプ列から
-- status を復元する。
UPDATE public.episodes SET status = 'deleted' WHERE deleted_at IS NOT NULL;
UPDATE public.episodes SET status = 'archived' WHERE unpublished_at IS NOT NULL AND deleted_at IS NULL;
