-- name: DeleteLibraryEntriesByUserID :one
-- ユーザーのライブラリエントリを最大batch_size件削除し、削除した件数を返す。
-- 見たい・見てる・見た (statuses.kindが1・2・3) のエントリの分だけ、同じ文の中で
-- works.watchers_countを減算する。1文にまとめることで、削除と減算が必ず一緒に
-- 確定し、再実行しても二重に減算しない。ステータスは後から削除するため、この時点では
-- 参照できる。updated_atはRailsのカウンタの更新と同じく変えない。
WITH deleted AS (
    DELETE FROM library_entries
    WHERE id IN (
        SELECT target.id
        FROM library_entries AS target
        WHERE target.user_id = sqlc.arg('user_id')
        LIMIT sqlc.arg('batch_size')
    )
    RETURNING library_entries.work_id, library_entries.status_id
),
watchers AS (
    SELECT deleted.work_id, COUNT(*) AS watchers_count
    FROM deleted
    INNER JOIN statuses ON statuses.id = deleted.status_id
    WHERE statuses.kind IN (1, 2, 3)
    GROUP BY deleted.work_id
),
decremented AS (
    UPDATE works
    SET watchers_count = works.watchers_count - watchers.watchers_count
    FROM watchers
    WHERE works.id = watchers.work_id
    RETURNING works.id
)
SELECT COUNT(*) FROM deleted;

-- name: UpdateLibraryEntryNextResources :one
-- 「見てる」のライブラリのうち、updated_from以降に更新され、次のエピソードが無いか、番組を
-- 選んでいて次の放送枠が無いものを、after_idより大きいidの順にbatch_size件読み、次の
-- エピソードと次の放送枠を求めて、値が変わる行だけを書き換える。読んだ行の最後のid
-- (読んだ行が無いときは0) と、書き換えた件数を返す。
-- 値の変わらない行は条件を満たしたまま残るため、ページは条件ではなくidで区切る。
-- 書き換える行は、読んだときとwatched_episode_ids・program_idが同じ行に限る。Railsの
-- 視聴記録が同じ行を同時に書き換えたときは、UPDATEが最新の行でこの条件を確かめ直すため、
-- 古い視聴済みのエピソードから求めた値で上書きせずに飛ばす。
WITH targets AS (
    SELECT
        library_entries.id,
        library_entries.watched_episode_ids,
        library_entries.program_id,
        next_episode.id AS next_episode_id,
        next_slot.id AS next_slot_id
    FROM library_entries
    INNER JOIN statuses ON statuses.id = library_entries.status_id
    LEFT JOIN LATERAL (
        SELECT episodes.id
        FROM episodes
        WHERE episodes.work_id = library_entries.work_id
            AND episodes.deleted_at IS NULL
            AND episodes.unpublished_at IS NULL
            AND NOT (episodes.id = ANY(library_entries.watched_episode_ids))
        ORDER BY episodes.sort_number, episodes.id
        LIMIT 1
    ) AS next_episode ON TRUE
    LEFT JOIN LATERAL (
        SELECT slots.id
        FROM slots
        WHERE slots.program_id = library_entries.program_id
            AND slots.episode_id = next_episode.id
            AND slots.deleted_at IS NULL
            AND slots.unpublished_at IS NULL
        ORDER BY slots.id
        LIMIT 1
    ) AS next_slot ON TRUE
    WHERE library_entries.id > sqlc.arg('after_id')
        AND statuses.kind = 2
        AND library_entries.updated_at >= sqlc.arg('updated_from')::timestamp
        AND (
            library_entries.next_episode_id IS NULL
            OR (library_entries.program_id IS NOT NULL AND library_entries.next_slot_id IS NULL)
        )
    ORDER BY library_entries.id
    LIMIT sqlc.arg('batch_size')
),
updated AS (
    UPDATE library_entries
    SET
        next_episode_id = targets.next_episode_id,
        next_slot_id = targets.next_slot_id,
        updated_at = sqlc.arg('now')::timestamp
    FROM targets
    WHERE library_entries.id = targets.id
        AND library_entries.watched_episode_ids = targets.watched_episode_ids
        AND library_entries.program_id IS NOT DISTINCT FROM targets.program_id
        AND (
            library_entries.next_episode_id IS DISTINCT FROM targets.next_episode_id
            OR library_entries.next_slot_id IS DISTINCT FROM targets.next_slot_id
        )
    RETURNING library_entries.id
)
SELECT
    COALESCE(MAX(targets.id), 0)::bigint AS last_id,
    (SELECT COUNT(*) FROM updated) AS updated_count
FROM targets;
