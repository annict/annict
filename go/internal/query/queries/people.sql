-- name: ListDBPeople :many
-- 公開中とアーカイブの人物を、IDの降順に並べる。archivedは閲覧を保つ凍結のため、誰にでも
-- 表示する。削除 (deleted) と掲載停止 (delisted) の人物は、名義を出さないため返さない。
SELECT
    p.id,
    p.name,
    p.name_kana,
    p.name_en,
    p.status
FROM people p
WHERE p.status IN ('published', 'archived')
ORDER BY p.id DESC
LIMIT sqlc.arg('per_page')
OFFSET sqlc.arg('page_offset')::bigint;

-- name: CountDBPeople :one
SELECT COUNT(*)
FROM people p
WHERE p.status IN ('published', 'archived');

-- name: ListTakenPersonNames :many
-- 一括登録の名前の重複確認に使う。people.nameの一意インデックス (index_people_on_name) と
-- 同じく、deletedの人物だけを除いて照合する。
SELECT p.name
FROM people p
WHERE p.name = ANY(sqlc.arg('names')::varchar[])
    AND p.status <> 'deleted';

-- name: CreatePerson :one
-- データ変更CTEで、Railsのsave_and_create_activity! と同じpeople.createの編集履歴を、
-- 挿入した行をparameters.newとして記録する (CreateEpisodeと同じ形)。人物は自身が
-- root_resourceになる (RailsのDbActivityMethods#root_resource)。
--
-- people・db_activitiesの時刻カラムはtimestamp without time zoneで、Rails版はUTCの時刻を
-- 書いている。NOW() をそのまま書くと接続のTimeZone設定で壁時計の値が変わるため、UTCに
-- 変換してから書く (UpdateDBPersonも同じ)。
WITH created_person AS (
    INSERT INTO people (
        name,
        name_kana,
        created_at,
        updated_at
    ) VALUES (
        sqlc.arg('name'),
        sqlc.arg('name_kana'),
        (NOW() AT TIME ZONE 'UTC'),
        (NOW() AT TIME ZONE 'UTC')
    )
    RETURNING *
), created_activity AS (
    INSERT INTO db_activities (
        user_id,
        trackable_id,
        trackable_type,
        action,
        parameters,
        created_at,
        updated_at,
        root_resource_id,
        root_resource_type
    )
    SELECT
        sqlc.arg('user_id'),
        cp.id,
        'Person',
        'people.create',
        json_build_object('new', row_to_json(cp)),
        (NOW() AT TIME ZONE 'UTC'),
        (NOW() AT TIME ZONE 'UTC'),
        cp.id,
        'Person'
    FROM created_person cp
    RETURNING id
)
SELECT cp.id
FROM created_person cp
CROSS JOIN created_activity ca;

-- name: GetPersonForEditByID :one
-- 編集フォームは、Rails版の編集フォームと同じ16項目と、hiddenで持ち回る版 (updated_at)
-- を読む。編集できるのは一覧に表示する公開中とアーカイブの人物だけで、削除と掲載停止の
-- 人物は除く。
SELECT
    p.id,
    p.name,
    p.name_kana,
    p.name_en,
    p.nickname,
    p.nickname_en,
    p.gender,
    p.blood_type,
    p.prefecture_id,
    p.birthday,
    p.height,
    p.url,
    p.url_en,
    p.wikipedia_url,
    p.wikipedia_url_en,
    p.twitter_username,
    p.twitter_username_en,
    p.status,
    p.updated_at
FROM people p
WHERE p.id = sqlc.arg('id')
    AND p.status IN ('published', 'archived');

-- name: ExistsPersonByNameExcludingID :one
-- 編集の名前の重複確認に使う。ListTakenPersonNamesと同じく、people.nameの一意インデックス
-- に揃えてdeletedの人物を除き、編集中の人物自身も除く。
SELECT EXISTS (
    SELECT 1
    FROM people p
    WHERE p.name = sqlc.arg('name')
        AND p.id <> sqlc.arg('exclude_id')
        AND p.status <> 'deleted'
);

-- name: UpdateDBPerson :one
-- 送信が名乗る版の照合は、UPDATEの中で行う (UpdateDBEpisodeと同じ)。people.updated_atは
-- timestamp without time zoneのため、版もtimestampにキャストして照合する。1行も返らない
-- ことは、フォームを開いた後に他者が行を書いたか、人物が編集できない状態 (削除・掲載停止)
-- になったことを意味する。
--
-- 編集履歴のpeople.updateは、内容が実際に変わったときにだけ記録する (Railsの
-- save_and_create_activity! と同じ)。NULL許容の文字列カラムは、Rails版が空欄を空文字列で、
-- Go版がNULLで保存するため、比較ではNULLIFで両者を同じ値とみなす。Rails版が保存した空文字列の
-- 欄を触らずに送信しただけで、変更として記録されないようにするため。parametersの形 (old /
-- newに行の全カラム) はRails版と揃え、/db/activitiesがそのまま差分を表示できるようにする。
-- created_activityは最後のSELECTにjoinしない。joinすると、変更の無い送信で返すべきidが
-- 落ちてしまうため。
--
-- Rails版は保存時にキャスト・スタッフのupdated_atも更新していた (touch_children) が、Go版
-- では再現しない。それらのupdated_atを読むキャッシュは無く、Rails版のtouchはキャストから作品の
-- updated_atまで連鎖するため、再現すると人物の編集が、開いている作品の編集フォームを競合させて
-- しまうため。
WITH current_person AS (
    SELECT p.*
    FROM people p
    WHERE p.id = sqlc.arg('id')
        AND p.status IN ('published', 'archived')
), updated_person AS (
    UPDATE people
    SET
        name = sqlc.arg('name'),
        name_kana = sqlc.arg('name_kana'),
        name_en = sqlc.arg('name_en'),
        nickname = sqlc.narg('nickname'),
        nickname_en = sqlc.arg('nickname_en'),
        gender = sqlc.narg('gender'),
        blood_type = sqlc.narg('blood_type'),
        prefecture_id = sqlc.narg('prefecture_id'),
        birthday = sqlc.narg('birthday'),
        height = sqlc.narg('height'),
        url = sqlc.narg('url'),
        url_en = sqlc.arg('url_en'),
        wikipedia_url = sqlc.narg('wikipedia_url'),
        wikipedia_url_en = sqlc.arg('wikipedia_url_en'),
        twitter_username = sqlc.narg('twitter_username'),
        twitter_username_en = sqlc.arg('twitter_username_en'),
        updated_at = (NOW() AT TIME ZONE 'UTC')
    FROM current_person cp
    WHERE people.id = cp.id
        AND people.status IN ('published', 'archived')
        AND people.updated_at IS NOT DISTINCT FROM sqlc.narg('version')::timestamp
    RETURNING people.*
), person_change AS (
    SELECT
        up.id,
        json_build_object('old', row_to_json(cp), 'new', row_to_json(up)) AS parameters
    FROM updated_person up
    JOIN current_person cp ON cp.id = up.id
    WHERE (
        cp.name,
        cp.name_kana,
        cp.name_en,
        NULLIF(cp.nickname, ''),
        cp.nickname_en,
        NULLIF(cp.gender, ''),
        NULLIF(cp.blood_type, ''),
        cp.prefecture_id,
        cp.birthday,
        cp.height,
        NULLIF(cp.url, ''),
        cp.url_en,
        NULLIF(cp.wikipedia_url, ''),
        cp.wikipedia_url_en,
        NULLIF(cp.twitter_username, ''),
        cp.twitter_username_en
    ) IS DISTINCT FROM (
        up.name,
        up.name_kana,
        up.name_en,
        NULLIF(up.nickname, ''),
        up.nickname_en,
        NULLIF(up.gender, ''),
        NULLIF(up.blood_type, ''),
        up.prefecture_id,
        up.birthday,
        up.height,
        NULLIF(up.url, ''),
        up.url_en,
        NULLIF(up.wikipedia_url, ''),
        up.wikipedia_url_en,
        NULLIF(up.twitter_username, ''),
        up.twitter_username_en
    )
), created_activity AS (
    INSERT INTO db_activities (
        user_id,
        trackable_id,
        trackable_type,
        action,
        parameters,
        created_at,
        updated_at,
        root_resource_id,
        root_resource_type
    )
    SELECT
        sqlc.arg('user_id'),
        pc.id,
        'Person',
        'people.update',
        pc.parameters,
        (NOW() AT TIME ZONE 'UTC'),
        (NOW() AT TIME ZONE 'UTC'),
        pc.id,
        'Person'
    FROM person_change pc
)
SELECT up.id
FROM updated_person up;

-- name: GetPersonForStateChangeByID :one
-- アーカイブ・アーカイブ解除・削除の確認画面が、対象の名前と現在の状態を読む。その画面が
-- 操作できない人物を呼び出し側で弾けるよう、状態で絞り込まずに返す
-- (GetWorkForStateChangeByIDと同じ)。
SELECT
    p.id,
    p.name,
    p.status
FROM people p
WHERE p.id = sqlc.arg('id');

-- name: ArchiveDBPerson :one
-- 公開中の人物をアーカイブする。archivedは閲覧できる状態のため、Rails版の表示から除外する
-- 橋渡し (unpublished_at) は書かない。状態の確認と書き込みを1文で行うため、確認画面を開いた
-- 後に他者が状態を変えた場合も、1行も返らない (not found) になる。時刻はCreatePersonと同じく
-- UTCに変換してから書く。
UPDATE people
SET
    status = 'archived',
    updated_at = (NOW() AT TIME ZONE 'UTC')
WHERE id = sqlc.arg('id')
    AND status = 'published'
RETURNING id;

-- name: UnarchiveDBPerson :one
-- アーカイブの人物を公開中に戻す。ArchiveDBPersonの逆で、アーカイブの人物だけを対象にする。
UPDATE people
SET
    status = 'published',
    updated_at = (NOW() AT TIME ZONE 'UTC')
WHERE id = sqlc.arg('id')
    AND status = 'archived'
RETURNING id;

-- name: DeleteDBPerson :one
-- 人物を論理削除する。Rails版の物理削除 (destroy_in_batches) と異なり、キャスト・スタッフ・
-- 編集履歴は残す。Rails版はperson_statusを読まないため、Rails版の表示から除外されるよう
-- deleted_atにも時刻を書く (Rails版への橋渡し)。削除できるのは一覧に表示する公開中と
-- アーカイブの人物で、削除済み・掲載停止の人物は対象にしない。
UPDATE people
SET
    status = 'deleted',
    deleted_at = (NOW() AT TIME ZONE 'UTC'),
    updated_at = (NOW() AT TIME ZONE 'UTC')
WHERE id = sqlc.arg('id')
    AND status IN ('published', 'archived')
RETURNING id;
