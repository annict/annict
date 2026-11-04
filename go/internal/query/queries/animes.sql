-- name: CreateAnime :one
INSERT INTO animes (
    title,
    title_kana,
    title_ro,
    title_en,
    title_alter,
    title_alter_ro,
    title_alter_en,
    title_alter_other,
    media,
    release_status,
    synopsis,
    synopsis_en,
    synopsis_source,
    synopsis_source_en,
    status,
    archive_message,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW()
) RETURNING *;

-- name: GetAnimeByID :one
SELECT * FROM animes
WHERE id = $1
LIMIT 1;

-- name: ListAnimesByIDs :many
SELECT * FROM animes
WHERE id = ANY($1::bigint[])
ORDER BY id;

-- name: UpdateAnime :exec
UPDATE animes
SET
    title = $2,
    title_kana = $3,
    title_ro = $4,
    title_en = $5,
    title_alter = $6,
    title_alter_ro = $7,
    title_alter_en = $8,
    title_alter_other = $9,
    media = $10,
    release_status = $11,
    synopsis = $12,
    synopsis_en = $13,
    synopsis_source = $14,
    synopsis_source_en = $15,
    status = $16,
    archive_message = $17,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateAnimeStatus :exec
-- 正本の行から導出したライフサイクル状態だけを更新する。非公開の経路はanime全体の
-- スナップショットを書き戻さず本クエリを使うことで、事前読み取り後にコミットされた無関係な
-- 内容編集を保持する。
UPDATE animes
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateAnimeTitle :exec
-- しょぼいカレンダーのサブタイトルをエピソードに入れたときに、そのエピソードのanimeへ同じ
-- タイトルを両書きする。ほかのカラムは書き戻さず、事前の読み取り後の編集を上書きしない。
UPDATE animes
SET
    title = sqlc.arg('title'),
    updated_at = NOW()
WHERE id = sqlc.arg('id');


-- REST API V2の作品の一覧は、並び順ごとにクエリを分ける。並び順をCASEで切り替えると、インデックスで並べられないため。
-- 絞り込みの条件は4つで同じにする。idsは候補のanime_id (年・季節の絞り込みは先にListAnimeIDsBySeasonsで候補にする)、
-- title_patternはLIKEの特殊文字をエスケープした部分一致のパターン。

-- name: ListAPIWorkIDsOrderByID :many
SELECT a.id
FROM animes a
INNER JOIN anime_classifications ac ON ac.anime_id = a.id
WHERE ac.kind = 'work' AND a.status = 'published'
    AND (sqlc.narg(ids)::bigint[] IS NULL OR a.id = ANY(sqlc.narg(ids)::bigint[]))
    AND (
        sqlc.narg(title_pattern)::text IS NULL
        OR a.title ILIKE sqlc.narg(title_pattern)::text
        OR a.title_kana ILIKE sqlc.narg(title_pattern)::text
    )
    AND (sqlc.narg(after_id)::bigint IS NULL OR a.id > sqlc.narg(after_id)::bigint)
ORDER BY a.id
LIMIT sqlc.arg(row_limit);

-- name: ListAPIWorkIDsOrderByIDDesc :many
SELECT a.id
FROM animes a
INNER JOIN anime_classifications ac ON ac.anime_id = a.id
WHERE ac.kind = 'work' AND a.status = 'published'
    AND (sqlc.narg(ids)::bigint[] IS NULL OR a.id = ANY(sqlc.narg(ids)::bigint[]))
    AND (
        sqlc.narg(title_pattern)::text IS NULL
        OR a.title ILIKE sqlc.narg(title_pattern)::text
        OR a.title_kana ILIKE sqlc.narg(title_pattern)::text
    )
    AND (sqlc.narg(after_id)::bigint IS NULL OR a.id < sqlc.narg(after_id)::bigint)
ORDER BY a.id DESC
LIMIT sqlc.arg(row_limit);

-- name: ListAPIWorkIDsOrderBySeason :many
-- 代表の季節のキー (model.AnimeSeason.SortKeyと同じ式)・anime_idの順に並べ、代表の季節の無い作品は最後に置く。
-- 代表の季節のある作品と無い作品を別々に並べて読み、季節のキーのインデックスで先頭から読めるようにする。
-- after_season_keyは前のページの最後の作品の季節のキーで、代表の季節の無い作品ならNULLにする
SELECT x.id
FROM (
    (
        SELECT a.id, 0 AS season_group, ps.season_key
        FROM (
            SELECT
                anime_id,
                year * 10 + CASE name
                    WHEN 'winter' THEN 1
                    WHEN 'spring' THEN 2
                    WHEN 'summer' THEN 3
                    WHEN 'fall' THEN 4
                    ELSE 5
                END AS season_key
            FROM anime_seasons
            WHERE is_primary
        ) ps
        INNER JOIN animes a ON a.id = ps.anime_id
        INNER JOIN anime_classifications ac ON ac.anime_id = a.id
        WHERE ac.kind = 'work' AND a.status = 'published'
            AND (sqlc.narg(ids)::bigint[] IS NULL OR a.id = ANY(sqlc.narg(ids)::bigint[]))
            AND (
                sqlc.narg(title_pattern)::text IS NULL
                OR a.title ILIKE sqlc.narg(title_pattern)::text
                OR a.title_kana ILIKE sqlc.narg(title_pattern)::text
            )
            AND (
                sqlc.narg(after_id)::bigint IS NULL
                OR ps.season_key > sqlc.narg(after_season_key)::integer
                OR (ps.season_key = sqlc.narg(after_season_key)::integer AND ps.anime_id > sqlc.narg(after_id)::bigint)
            )
        ORDER BY ps.season_key ASC, ps.anime_id ASC
        LIMIT sqlc.arg(row_limit)
    )
    UNION ALL
    (
        SELECT a.id, 1 AS season_group, NULL::integer AS season_key
        FROM animes a
        INNER JOIN anime_classifications ac ON ac.anime_id = a.id
        WHERE ac.kind = 'work' AND a.status = 'published'
            AND NOT EXISTS (SELECT 1 FROM anime_seasons ps WHERE ps.anime_id = a.id AND ps.is_primary)
            AND (sqlc.narg(ids)::bigint[] IS NULL OR a.id = ANY(sqlc.narg(ids)::bigint[]))
            AND (
                sqlc.narg(title_pattern)::text IS NULL
                OR a.title ILIKE sqlc.narg(title_pattern)::text
                OR a.title_kana ILIKE sqlc.narg(title_pattern)::text
            )
            AND (
                sqlc.narg(after_id)::bigint IS NULL
                OR sqlc.narg(after_season_key)::integer IS NOT NULL
                OR a.id > sqlc.narg(after_id)::bigint
            )
        ORDER BY a.id ASC
        LIMIT sqlc.arg(row_limit)
    )
) x
ORDER BY x.season_group, x.season_key ASC, x.id ASC
LIMIT sqlc.arg(row_limit);

-- name: ListAPIWorkIDsOrderBySeasonDesc :many
-- 季節の新しい順に並べる。代表の季節の無い作品は、こちらでも最後に置く
SELECT x.id
FROM (
    (
        SELECT a.id, 0 AS season_group, ps.season_key
        FROM (
            SELECT
                anime_id,
                year * 10 + CASE name
                    WHEN 'winter' THEN 1
                    WHEN 'spring' THEN 2
                    WHEN 'summer' THEN 3
                    WHEN 'fall' THEN 4
                    ELSE 5
                END AS season_key
            FROM anime_seasons
            WHERE is_primary
        ) ps
        INNER JOIN animes a ON a.id = ps.anime_id
        INNER JOIN anime_classifications ac ON ac.anime_id = a.id
        WHERE ac.kind = 'work' AND a.status = 'published'
            AND (sqlc.narg(ids)::bigint[] IS NULL OR a.id = ANY(sqlc.narg(ids)::bigint[]))
            AND (
                sqlc.narg(title_pattern)::text IS NULL
                OR a.title ILIKE sqlc.narg(title_pattern)::text
                OR a.title_kana ILIKE sqlc.narg(title_pattern)::text
            )
            AND (
                sqlc.narg(after_id)::bigint IS NULL
                OR ps.season_key < sqlc.narg(after_season_key)::integer
                OR (ps.season_key = sqlc.narg(after_season_key)::integer AND ps.anime_id < sqlc.narg(after_id)::bigint)
            )
        ORDER BY ps.season_key DESC, ps.anime_id DESC
        LIMIT sqlc.arg(row_limit)
    )
    UNION ALL
    (
        SELECT a.id, 1 AS season_group, NULL::integer AS season_key
        FROM animes a
        INNER JOIN anime_classifications ac ON ac.anime_id = a.id
        WHERE ac.kind = 'work' AND a.status = 'published'
            AND NOT EXISTS (SELECT 1 FROM anime_seasons ps WHERE ps.anime_id = a.id AND ps.is_primary)
            AND (sqlc.narg(ids)::bigint[] IS NULL OR a.id = ANY(sqlc.narg(ids)::bigint[]))
            AND (
                sqlc.narg(title_pattern)::text IS NULL
                OR a.title ILIKE sqlc.narg(title_pattern)::text
                OR a.title_kana ILIKE sqlc.narg(title_pattern)::text
            )
            AND (
                sqlc.narg(after_id)::bigint IS NULL
                OR sqlc.narg(after_season_key)::integer IS NOT NULL
                OR a.id < sqlc.narg(after_id)::bigint
            )
        ORDER BY a.id DESC
        LIMIT sqlc.arg(row_limit)
    )
) x
ORDER BY x.season_group, x.season_key DESC, x.id DESC
LIMIT sqlc.arg(row_limit);

-- name: ListAPIEpisodeIDs :many
-- 作品の公開中のエピソードを (sort_number, anime_id) の順に並べる
SELECT ac.anime_id
FROM anime_classifications ac
INNER JOIN animes a ON a.id = ac.anime_id
WHERE ac.parent_anime_id = sqlc.arg(parent_anime_id)::bigint
    AND a.status = 'published'
    AND (
        sqlc.narg(after_anime_id)::bigint IS NULL
        OR ac.sort_number > sqlc.narg(after_sort_number)::integer
        OR (ac.sort_number = sqlc.narg(after_sort_number)::integer AND ac.anime_id > sqlc.narg(after_anime_id)::bigint)
    )
ORDER BY ac.sort_number, ac.anime_id
LIMIT sqlc.arg(row_limit);
