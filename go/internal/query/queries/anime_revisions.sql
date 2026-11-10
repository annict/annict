-- name: CreateAnimeRevision :exec
-- 操作に属する作品・エピソードの版を1件作る。
INSERT INTO anime_revisions (db_edit_id, anime_id, parent_anime_id, snapshot, snapshot_version)
VALUES (sqlc.arg('db_edit_id'), sqlc.arg('anime_id'), sqlc.narg('parent_anime_id'), sqlc.arg('snapshot'), sqlc.arg('snapshot_version'));

-- name: CreateInitialAnimeRevision :exec
-- 起点の版 (db_edit_idがNULLの版) を1件作る。同じ作品・エピソードの初めての記録が同時に走ったときは、
-- 部分一意インデックスで先に作られた起点の版を残し、この行は足さない。
INSERT INTO anime_revisions (db_edit_id, anime_id, parent_anime_id, snapshot, snapshot_version)
VALUES (NULL, sqlc.arg('anime_id'), sqlc.narg('parent_anime_id'), sqlc.arg('snapshot'), sqlc.arg('snapshot_version'))
ON CONFLICT (anime_id) WHERE db_edit_id IS NULL DO NOTHING;

-- name: ListAnimeIDsWithRevisions :many
-- 指定した作品・エピソードのうち、版を1つ以上持つもののIDを返す。
SELECT a.id::bigint AS anime_id
FROM unnest(sqlc.arg('anime_ids')::bigint[]) AS a(id)
WHERE EXISTS (
    SELECT 1
    FROM anime_revisions ar
    WHERE ar.anime_id = a.id
);

-- name: ListAnimeRevisionsByAnimeID :many
-- 作品・エピソードの版を、起点の版を含めてIDの昇順に返す。
SELECT id, db_edit_id, anime_id, snapshot, snapshot_version, created_at
FROM anime_revisions
WHERE anime_id = sqlc.arg('anime_id')
ORDER BY id;

-- name: ListAnimeRevisionHistory :many
-- 作品・エピソードの操作の版を、新しい順に1ページ分返す。各版には、起点の版を含めて同じ作品・エピソードの
-- 1つ前の版のスナップショットと、操作・編集者のアットネームを付ける。1つ前の版は差分の比べる相手で、
-- ページの外にあっても取れるよう、起点の版を含めたすべての版に対してLAGを取ってからページを切る。
SELECT
    r.id,
    r.db_edit_id::bigint AS db_edit_id,
    r.anime_id,
    r.snapshot,
    r.snapshot_version,
    r.created_at,
    -- 1つ前の版が無いときは、版を0、スナップショットをJSONのnullにして返す。
    COALESCE(r.previous_snapshot, 'null'::jsonb)::jsonb AS previous_snapshot,
    COALESCE(r.previous_snapshot_version, 0)::smallint AS previous_snapshot_version,
    de.actor,
    de.user_id,
    de.action,
    de.created_at AS db_edit_created_at,
    -- 編集者のいない操作 (システムの書き込みと退会したユーザーの編集) では空文字列にする。
    COALESCE(u.username, '')::varchar AS username,
    -- 編集者のアバターに使う。編集者のいない操作とプロフィール画像の無い編集者ではNULLになる。
    p.image_data AS profile_image_data
FROM (
    SELECT
        ar.*,
        LAG(ar.snapshot) OVER w AS previous_snapshot,
        LAG(ar.snapshot_version) OVER w AS previous_snapshot_version
    FROM anime_revisions ar
    WHERE ar.anime_id = sqlc.arg('anime_id')
    WINDOW w AS (ORDER BY ar.id)
) r
JOIN db_edits de ON de.id = r.db_edit_id
LEFT JOIN users u ON u.id = de.user_id
LEFT JOIN profiles p ON p.user_id = u.id
ORDER BY r.id DESC
LIMIT sqlc.arg('per_page')
OFFSET sqlc.arg('page_offset')::bigint;

-- name: CountAnimeRevisionHistory :one
-- 作品・エピソードの操作の版の数を返す。起点の版は履歴に並べないため数えない。
SELECT COUNT(*)
FROM anime_revisions
WHERE anime_id = sqlc.arg('anime_id')
    AND db_edit_id IS NOT NULL;

-- name: ListWorkAnimeRevisionHistory :many
-- 作品の操作の版と、版の時点でその作品の配下にあったエピソードの操作の版を、新しい順に1ページ分返す。
-- 1つ前の版はエピソードごとに取るため、作品とエピソードのすべての版に対して、作品・エピソードで分けたLAGを
-- 取ってからページを切る。削除したエピソードの版は、Annict DBのエピソードの一覧と同じく誰にも見せないため除く。
-- 版の時点と今の親作品が異なる場合も、今の親作品が削除されていればエピソードの版を除く。
-- エピソードの版には、履歴の画面へのリンクに使うepisodes.idを付け、作品の版では0にする。
SELECT
    r.id,
    r.db_edit_id::bigint AS db_edit_id,
    r.anime_id,
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
    p.image_data AS profile_image_data,
    COALESCE(e.id, 0)::bigint AS episode_id
FROM (
    SELECT
        ar.*,
        LAG(ar.snapshot) OVER w AS previous_snapshot,
        LAG(ar.snapshot_version) OVER w AS previous_snapshot_version
    FROM anime_revisions ar
    WHERE ar.anime_id IN (
        SELECT tr.anime_id
        FROM anime_revisions tr
        WHERE tr.anime_id = sqlc.arg('anime_id')
            OR tr.parent_anime_id = sqlc.arg('anime_id')
    )
    WINDOW w AS (PARTITION BY ar.anime_id ORDER BY ar.id)
) r
JOIN db_edits de ON de.id = r.db_edit_id
LEFT JOIN users u ON u.id = de.user_id
LEFT JOIN profiles p ON p.user_id = u.id
LEFT JOIN episodes e ON e.anime_id = r.anime_id AND r.anime_id <> sqlc.arg('anime_id')
WHERE r.anime_id = sqlc.arg('anime_id')
    OR (
        r.parent_anime_id = sqlc.arg('anime_id')
        AND e.id IS NOT NULL
        AND e.deleted_at IS NULL
        AND EXISTS (SELECT 1 FROM works w WHERE w.id = e.work_id AND w.deleted_at IS NULL)
    )
ORDER BY r.id DESC
LIMIT sqlc.arg('per_page')
OFFSET sqlc.arg('page_offset')::bigint;

-- name: CountWorkAnimeRevisionHistory :one
-- ListWorkAnimeRevisionHistoryが並べる版の数を返す。
SELECT COUNT(*)
FROM anime_revisions r
LEFT JOIN episodes e ON e.anime_id = r.anime_id AND r.anime_id <> sqlc.arg('anime_id')
WHERE r.db_edit_id IS NOT NULL
    AND (
        r.anime_id = sqlc.arg('anime_id')
        OR (
            r.parent_anime_id = sqlc.arg('anime_id')
            AND e.id IS NOT NULL
            AND e.deleted_at IS NULL
            AND EXISTS (SELECT 1 FROM works w WHERE w.id = e.work_id AND w.deleted_at IS NULL)
        )
    );

-- name: ListAnimeRevisionsForDBEdits :many
-- Annict DBの編集履歴の1ページ分の操作に属する、作品・エピソードの版をIDの昇順に返す。各版には、起点の版を
-- 含めた同じ作品・エピソードの1つ前の版と、今のリソースへのリンクに使うworks.id・episodes.id (無いときは0) を
-- 付ける。work_idとwork_titleは、作品の版ではその作品、エピソードの版では今の親作品のもの。今のリソースが
-- Annict DBの一覧で見えない版 (削除した作品・エピソード) は除く。
SELECT
    r.id,
    r.db_edit_id::bigint AS db_edit_id,
    r.anime_id,
    r.snapshot,
    r.snapshot_version,
    r.created_at,
    COALESCE(r.previous_snapshot, 'null'::jsonb)::jsonb AS previous_snapshot,
    COALESCE(r.previous_snapshot_version, 0)::smallint AS previous_snapshot_version,
    COALESCE(w.id, ew.id, 0)::bigint AS work_id,
    COALESCE(w.title, ew.title, '')::varchar AS work_title,
    COALESCE(e.id, 0)::bigint AS episode_id
FROM (
    SELECT
        ar.*,
        LAG(ar.snapshot) OVER w AS previous_snapshot,
        LAG(ar.snapshot_version) OVER w AS previous_snapshot_version
    FROM anime_revisions ar
    WHERE ar.anime_id IN (
        SELECT tr.anime_id
        FROM anime_revisions tr
        WHERE tr.db_edit_id = ANY(sqlc.arg('db_edit_ids')::bigint[])
    )
    WINDOW w AS (PARTITION BY ar.anime_id ORDER BY ar.id)
) r
LEFT JOIN works w ON w.anime_id = r.anime_id AND w.deleted_at IS NULL
LEFT JOIN episodes e ON e.anime_id = r.anime_id
    AND e.deleted_at IS NULL
    AND EXISTS (SELECT 1 FROM works pw WHERE pw.id = e.work_id AND pw.deleted_at IS NULL)
LEFT JOIN works ew ON ew.id = e.work_id
WHERE r.db_edit_id = ANY(sqlc.arg('db_edit_ids')::bigint[])
    AND (w.id IS NOT NULL OR e.id IS NOT NULL)
ORDER BY r.id;
