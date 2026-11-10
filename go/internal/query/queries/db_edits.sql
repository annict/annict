-- name: CreateDBEdit :one
-- Annict DBの編集履歴の操作を1件作る。created_atはtimestamp with time zoneのため、NOW() を
-- そのまま書く。
INSERT INTO db_edits (actor, user_id, action)
VALUES (sqlc.arg('actor'), sqlc.narg('user_id'), sqlc.arg('action'))
RETURNING id, actor, user_id, action, created_at;

-- name: ListDBEdits :many
-- Annict DBの編集履歴の操作を、IDの降順に1ページ分返す。見える版を1つも持たない操作は、ページの件数が
-- 欠けないよう、ページを切る前に除く。編集者のいない操作 (システムの書き込みと退会したユーザーの編集) では
-- アットネームを空文字列にする。
SELECT
    de.id,
    de.actor,
    de.user_id,
    de.action,
    de.created_at,
    COALESCE(u.username, '')::varchar AS username,
    -- 編集者のアバターに使う。編集者のいない操作とプロフィール画像の無い編集者ではNULLになる。
    pf.image_data AS profile_image_data
FROM db_edits de
LEFT JOIN users u ON u.id = de.user_id
LEFT JOIN profiles pf ON pf.user_id = u.id
WHERE EXISTS (
    -- 作品・エピソードの版は、今のリソースがAnnict DBの一覧で見えるとき (削除されていない作品と、
    -- 削除されていない作品の削除されていないエピソード) だけ見せる。
    SELECT 1
    FROM anime_revisions ar
    WHERE ar.db_edit_id = de.id
        AND (
            EXISTS (SELECT 1 FROM works w WHERE w.anime_id = ar.anime_id AND w.deleted_at IS NULL)
            OR EXISTS (
                SELECT 1
                FROM episodes e
                JOIN works ew ON ew.id = e.work_id
                WHERE e.anime_id = ar.anime_id
                    AND e.deleted_at IS NULL
                    AND ew.deleted_at IS NULL
            )
        )
) OR EXISTS (
    -- 人物の版は、削除した人物では見せない。管理者以外には、掲載停止中の人物の版と、掲載停止・
    -- 掲載再開の操作の版と、掲載停止中に記録した版を見せない (人物の履歴と同じ)。
    SELECT 1
    FROM person_revisions pr
    JOIN people p ON p.id = pr.person_id
    WHERE pr.db_edit_id = de.id
        AND p.status <> 'deleted'
        AND (
            sqlc.arg('include_delisted')::boolean
            OR (
                p.status <> 'delisted'
                AND de.action NOT IN ('delist', 'undelist')
                AND pr.snapshot->>'status' <> 'delisted'
            )
        )
)
ORDER BY de.id DESC
LIMIT sqlc.arg('per_page')
OFFSET sqlc.arg('page_offset')::bigint;

-- name: CountDBEdits :one
-- ListDBEditsが並べる操作の数を返す。
SELECT COUNT(*)
FROM db_edits de
WHERE EXISTS (
    -- 作品・エピソードの版は、今のリソースがAnnict DBの一覧で見えるとき (削除されていない作品と、
    -- 削除されていない作品の削除されていないエピソード) だけ見せる。
    SELECT 1
    FROM anime_revisions ar
    WHERE ar.db_edit_id = de.id
        AND (
            EXISTS (SELECT 1 FROM works w WHERE w.anime_id = ar.anime_id AND w.deleted_at IS NULL)
            OR EXISTS (
                SELECT 1
                FROM episodes e
                JOIN works ew ON ew.id = e.work_id
                WHERE e.anime_id = ar.anime_id
                    AND e.deleted_at IS NULL
                    AND ew.deleted_at IS NULL
            )
        )
) OR EXISTS (
    -- 人物の版は、削除した人物では見せない。管理者以外には、掲載停止中の人物の版と、掲載停止・
    -- 掲載再開の操作の版と、掲載停止中に記録した版を見せない (人物の履歴と同じ)。
    SELECT 1
    FROM person_revisions pr
    JOIN people p ON p.id = pr.person_id
    WHERE pr.db_edit_id = de.id
        AND p.status <> 'deleted'
        AND (
            sqlc.arg('include_delisted')::boolean
            OR (
                p.status <> 'delisted'
                AND de.action NOT IN ('delist', 'undelist')
                AND pr.snapshot->>'status' <> 'delisted'
            )
        )
);

-- name: DetachDBEditsFromUser :one
-- ユーザーが行ったAnnict DBの編集の操作を最大batch_size件、編集者との紐づきを外し、外した件数を
-- 返す。前後の版の差分が正しく出るよう、操作と版は残す。
WITH detached AS (
    UPDATE db_edits
    SET user_id = NULL
    WHERE id IN (
        SELECT target.id
        FROM db_edits AS target
        WHERE target.user_id = sqlc.arg('user_id')::bigint
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING db_edits.id
)
SELECT COUNT(*) FROM detached;
