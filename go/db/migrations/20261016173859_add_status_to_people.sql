-- migrate:up

-- people (約3万件) はRails版と共有する書き込み中のテーブルのため、ロックを待ち続けて後続の
-- 読み書きを詰まらせないよう、ロックを取れなければすぐに失敗させる。定数の既定値を持つ
-- ADD COLUMNは行を書き換えず、状態の移し替えと一意インデックスの作り直しも数千行と3万行に
-- とどまるため、トランザクション全体のロック時間は短い。
SET LOCAL lock_timeout = '5s';

-- 人物の状態を、unpublished_at / deleted_atの有無ではなく単一のenumで宣言する
-- (animes.statusと同じ語彙)。delistedは本人の申出による掲載停止で、物理削除して構わない
-- deletedとは別の値にする。
CREATE TYPE public.person_status AS ENUM ('published', 'archived', 'delisted', 'deleted');

-- archive_messageはarchivedの人物ページに表示する案内文、delist_messageは管理者だけが参照する
-- 掲載停止の対応記録。
ALTER TABLE public.people
    ADD COLUMN status public.person_status DEFAULT 'published' NOT NULL,
    ADD COLUMN archive_message character varying,
    ADD COLUMN delist_message character varying;

-- 既存の状態をstatusに移す。移した後のunpublished_atは、Rails版の表示から人物を除外する
-- 掲載停止の橋渡しにだけ使うため、移し替えた行ではNULLに戻す。Rails版はdeleted_atでも人物を
-- 除外するため、deletedに移す行はdeleted_atを立てたままにする。

-- 削除済みの人物 (aasm_stateがhiddenの行は必ずdeleted_atを伴う)。
UPDATE public.people
SET
    status = 'deleted',
    unpublished_at = NULL
WHERE deleted_at IS NOT NULL;

-- 非公開の人物のうち、お気に入りが付いている人物。archivedは閲覧できる状態のため、
-- お気に入りを付けた利用者がその人物のページに再び到達できる。
UPDATE public.people p
SET
    status = 'archived',
    unpublished_at = NULL
WHERE p.deleted_at IS NULL
    AND p.unpublished_at IS NOT NULL
    AND EXISTS (
        SELECT 1
        FROM public.person_favorites pf
        WHERE pf.person_id = p.id
    );

-- 残りの非公開の人物。大半は編集者が整理した古い一括登録の人物で、閲覧できる状態に戻すと
-- 中身の無い人物ページが公開されるため削除に移す。Rails版の表示から引き続き除外されるよう、
-- 非公開にした時刻をdeleted_atに移す。
UPDATE public.people
SET
    status = 'deleted',
    deleted_at = unpublished_at,
    unpublished_at = NULL
WHERE deleted_at IS NULL
    AND unpublished_at IS NOT NULL;

-- 削除した人物が名前を占有し続けないよう、名前の一意性の対象からdeletedの人物を外す。
-- 掲載停止 (delisted) の人物は対象に残り、同じ名前での再登録を防ぐ。インデックス名はGo版が
-- 一意性違反の判定に使うため変えない。
DROP INDEX public.index_people_on_name;
CREATE UNIQUE INDEX index_people_on_name ON public.people USING btree (name) WHERE (status <> 'deleted'::public.person_status);

-- migrate:down

SET LOCAL lock_timeout = '5s';

-- deletedの人物と同じ名前の人物が登録されていると、全行を対象にした一意インデックスは
-- 作れずにロールバックが失敗する。その場合はどちらかの人物の名前を変えてから戻す。
DROP INDEX public.index_people_on_name;
CREATE UNIQUE INDEX index_people_on_name ON public.people USING btree (name);

-- archivedの人物はunpublished_atをNULLに戻しているため、非公開として扱われるよう時刻を書き戻す。
-- deletedの人物はdeleted_at、delistedの人物はunpublished_atを立てたままにしているため、
-- 書き戻しは要らない。
UPDATE public.people
SET unpublished_at = (NOW() AT TIME ZONE 'UTC')
WHERE status = 'archived'
    AND unpublished_at IS NULL;

ALTER TABLE public.people
    DROP COLUMN delist_message,
    DROP COLUMN archive_message,
    DROP COLUMN status;

DROP TYPE public.person_status;
