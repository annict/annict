-- name: CreatePersonRevision :exec
-- 操作に属する人物の版を1件作る。
INSERT INTO person_revisions (db_edit_id, person_id, snapshot, snapshot_version)
VALUES (sqlc.arg('db_edit_id'), sqlc.arg('person_id'), sqlc.arg('snapshot'), sqlc.arg('snapshot_version'));

-- name: CreateInitialPersonRevision :exec
-- 人物の起点の版を1件作る。CreateInitialAnimeRevisionと同じく、先に作られた起点の版があれば足さない。
INSERT INTO person_revisions (db_edit_id, person_id, snapshot, snapshot_version)
VALUES (NULL, sqlc.arg('person_id'), sqlc.arg('snapshot'), sqlc.arg('snapshot_version'))
ON CONFLICT (person_id) WHERE db_edit_id IS NULL DO NOTHING;

-- name: ListPersonIDsWithRevisions :many
-- 指定した人物のうち、版を1つ以上持つもののIDを返す。
SELECT p.id::bigint AS person_id
FROM unnest(sqlc.arg('person_ids')::bigint[]) AS p(id)
WHERE EXISTS (
    SELECT 1
    FROM person_revisions pr
    WHERE pr.person_id = p.id
);

-- name: ListPersonRevisionsByPersonID :many
-- 人物の版を、起点の版を含めてIDの昇順に返す。
SELECT id, db_edit_id, person_id, snapshot, snapshot_version, created_at
FROM person_revisions
WHERE person_id = sqlc.arg('person_id')
ORDER BY id;

-- name: ListPersonRevisionHistory :many
-- 人物の操作の版を、新しい順に1ページ分返す。各版には、起点の版と隠す版を含めた1つ前の版のスナップショットと、
-- 操作・編集者のアットネームを付ける。include_delistedが偽のときは、掲載停止・掲載再開の操作の版と、
-- 掲載停止中に記録した版 (スナップショットのstatusがdelisted) を除く。掲載停止の事実を管理者以外に
-- 伝えないため。1つ前の版は隠す版も含めてLAGを取ってからページを切るため、見せる版の差分に隠した版の変更は出ない。
SELECT
    r.id,
    r.db_edit_id::bigint AS db_edit_id,
    r.person_id,
    r.snapshot,
    r.snapshot_version,
    r.created_at,
    COALESCE(r.previous_snapshot, 'null'::jsonb)::jsonb AS previous_snapshot,
    COALESCE(r.previous_snapshot_version, 0)::smallint AS previous_snapshot_version,
    de.actor,
    de.user_id,
    de.action,
    de.created_at AS db_edit_created_at,
    COALESCE(u.username, '')::varchar AS username,
    -- 編集者のアバターに使う。編集者のいない操作とプロフィール画像の無い編集者ではNULLになる。
    p.image_data AS profile_image_data
FROM (
    SELECT
        pr.*,
        LAG(pr.snapshot) OVER w AS previous_snapshot,
        LAG(pr.snapshot_version) OVER w AS previous_snapshot_version
    FROM person_revisions pr
    WHERE pr.person_id = sqlc.arg('person_id')
    WINDOW w AS (ORDER BY pr.id)
) r
JOIN db_edits de ON de.id = r.db_edit_id
LEFT JOIN users u ON u.id = de.user_id
LEFT JOIN profiles p ON p.user_id = u.id
WHERE sqlc.arg('include_delisted')::boolean
    OR (de.action NOT IN ('delist', 'undelist') AND r.snapshot->>'status' <> 'delisted')
ORDER BY r.id DESC
LIMIT sqlc.arg('per_page')
OFFSET sqlc.arg('page_offset')::bigint;

-- name: CountPersonRevisionHistory :one
-- ListPersonRevisionHistoryが並べる版の数を返す。
SELECT COUNT(*)
FROM person_revisions r
JOIN db_edits de ON de.id = r.db_edit_id
WHERE r.person_id = sqlc.arg('person_id')
    AND (
        sqlc.arg('include_delisted')::boolean
        OR (de.action NOT IN ('delist', 'undelist') AND r.snapshot->>'status' <> 'delisted')
    );

-- name: ListPersonRevisionsForDBEdits :many
-- Annict DBの編集履歴の1ページ分の操作に属する、人物の版をIDの昇順に返す。各版には、起点の版と隠す版を含めた
-- 同じ人物の1つ前の版を付ける。見せる版の条件はListDBEditsと同じ。
SELECT
    r.id,
    r.db_edit_id::bigint AS db_edit_id,
    r.person_id,
    r.snapshot,
    r.snapshot_version,
    r.created_at,
    COALESCE(r.previous_snapshot, 'null'::jsonb)::jsonb AS previous_snapshot,
    COALESCE(r.previous_snapshot_version, 0)::smallint AS previous_snapshot_version
FROM (
    SELECT
        pr.*,
        LAG(pr.snapshot) OVER w AS previous_snapshot,
        LAG(pr.snapshot_version) OVER w AS previous_snapshot_version
    FROM person_revisions pr
    WHERE pr.person_id IN (
        SELECT tr.person_id
        FROM person_revisions tr
        WHERE tr.db_edit_id = ANY(sqlc.arg('db_edit_ids')::bigint[])
    )
    WINDOW w AS (PARTITION BY pr.person_id ORDER BY pr.id)
) r
JOIN db_edits de ON de.id = r.db_edit_id
JOIN people p ON p.id = r.person_id
WHERE r.db_edit_id = ANY(sqlc.arg('db_edit_ids')::bigint[])
    AND p.status <> 'deleted'
    AND (
        sqlc.arg('include_delisted')::boolean
        OR (
            p.status <> 'delisted'
            AND de.action NOT IN ('delist', 'undelist')
            AND r.snapshot->>'status' <> 'delisted'
        )
    )
ORDER BY r.id;
