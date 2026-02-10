-- migrate:up

-- エピソードの状態を表す enum 型を作成
CREATE TYPE public.episode_status AS ENUM ('published', 'archived', 'deleted');

-- episodes テーブルに status と archive_message カラムを追加
ALTER TABLE public.episodes ADD COLUMN status public.episode_status NOT NULL DEFAULT 'published';
ALTER TABLE public.episodes ADD COLUMN archive_message VARCHAR;

-- 公開中のエピソードの検索用インデックス
CREATE INDEX index_episodes_on_status ON public.episodes(status) WHERE status = 'published';

-- 既存データの移行
UPDATE public.episodes SET status = 'deleted' WHERE deleted_at IS NOT NULL;
UPDATE public.episodes SET status = 'archived' WHERE unpublished_at IS NOT NULL AND deleted_at IS NULL;

-- migrate:down

-- インデックスの削除
DROP INDEX IF EXISTS public.index_episodes_on_status;

-- カラムの削除
ALTER TABLE public.episodes DROP COLUMN IF EXISTS archive_message;
ALTER TABLE public.episodes DROP COLUMN IF EXISTS status;

-- enum 型の削除
DROP TYPE IF EXISTS public.episode_status;
