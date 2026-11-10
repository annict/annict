-- migrate:up

-- Annict DBの編集履歴を、1回の保存を表す操作 (db_edits) と、リソースごとの版
-- (anime_revisions / person_revisions) の2段で持つ。版は外部キーで対象を指し、変更後の状態を
-- Go版で定義した形のスナップショットとして持つ。どれも新しいテーブルで、既存のテーブルは変えない。

-- db_edit_actorは編集した主体。userは編集者の編集、systemはエピソードの自動生成のような
-- アプリケーションの書き込み。
CREATE TYPE public.db_edit_actor AS ENUM ('user', 'system');

-- db_edit_actionは操作の種類。delist / undelistは人物の掲載停止と掲載再開。
CREATE TYPE public.db_edit_action AS ENUM ('create', 'update', 'archive', 'unarchive', 'delete', 'delist', 'undelist');

CREATE TABLE public.db_edits (
    id BIGSERIAL PRIMARY KEY,
    actor public.db_edit_actor NOT NULL,
    -- 編集者。退会するとNULLにして編集者を外し、版は残す。システムの書き込みは持たない。
    user_id BIGINT REFERENCES public.users(id) ON DELETE SET NULL,
    action public.db_edit_action NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT db_edits_user_id_check CHECK (actor = 'user' OR user_id IS NULL)
);

-- 退会処理が編集者を外す行を探すため、外部キーのSET NULLのため。
CREATE INDEX index_db_edits_on_user_id ON public.db_edits(user_id);

-- 作品・エピソードの版。版は操作に従属する記録のため、操作とリソースの削除に従って消す。
CREATE TABLE public.anime_revisions (
    id BIGSERIAL PRIMARY KEY,
    -- NULLの行は起点の版 (記録を始めた時点の状態)。
    db_edit_id BIGINT REFERENCES public.db_edits(id) ON DELETE CASCADE,
    anime_id BIGINT NOT NULL REFERENCES public.animes(id) ON DELETE CASCADE,
    -- 版の時点の親の作品 (エピソードの版のとき)。
    parent_anime_id BIGINT REFERENCES public.animes(id) ON DELETE SET NULL,
    snapshot JSONB NOT NULL,
    snapshot_version SMALLINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 1回の操作でリソースごとに版を1つにする。外部キーのCASCADEのための索引も兼ねる。
CREATE UNIQUE INDEX index_anime_revisions_on_db_edit_id_and_anime_id ON public.anime_revisions(db_edit_id, anime_id);
-- 同じリソースの初めての記録が同時に走っても、起点の版を1つに保つ。
CREATE UNIQUE INDEX index_anime_revisions_on_anime_id_initial ON public.anime_revisions(anime_id) WHERE db_edit_id IS NULL;
CREATE INDEX index_anime_revisions_on_anime_id_and_id ON public.anime_revisions(anime_id, id);
CREATE INDEX index_anime_revisions_on_parent_anime_id_and_id ON public.anime_revisions(parent_anime_id, id);

-- 人物の版。anime_revisionsと同じ形で、親の列を持たない。
CREATE TABLE public.person_revisions (
    id BIGSERIAL PRIMARY KEY,
    db_edit_id BIGINT REFERENCES public.db_edits(id) ON DELETE CASCADE,
    person_id BIGINT NOT NULL REFERENCES public.people(id) ON DELETE CASCADE,
    snapshot JSONB NOT NULL,
    snapshot_version SMALLINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX index_person_revisions_on_db_edit_id_and_person_id ON public.person_revisions(db_edit_id, person_id);
CREATE UNIQUE INDEX index_person_revisions_on_person_id_initial ON public.person_revisions(person_id) WHERE db_edit_id IS NULL;
CREATE INDEX index_person_revisions_on_person_id_and_id ON public.person_revisions(person_id, id);

-- migrate:down

DROP TABLE IF EXISTS public.person_revisions;
DROP TABLE IF EXISTS public.anime_revisions;
DROP TABLE IF EXISTS public.db_edits;

DROP TYPE IF EXISTS public.db_edit_action;
DROP TYPE IF EXISTS public.db_edit_actor;
