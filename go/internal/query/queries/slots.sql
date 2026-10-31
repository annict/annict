-- name: ListSlotsForEpisodeGeneration :many
-- 自動生成の対象はRailsのepisode:generateと同じ条件 (未削除・公開中で、エピソードが紐づいて
-- おらず、再放送でも不定期放送でもなく、番組と回数を持つ) の放送枠に、開始日時の範囲と作品が
-- 未削除・公開中であることを加えたもの。
-- 一覧では処理するidと所属作品だけを読み、話数の計算に使う値は作品ロック後に読み直す。
SELECT s.id, s.work_id
FROM slots s
JOIN works w ON w.id = s.work_id
WHERE s.deleted_at IS NULL
    AND s.unpublished_at IS NULL
    AND s.episode_id IS NULL
    AND NOT s.rebroadcast
    AND NOT s.irregular
    AND s.program_id IS NOT NULL
    AND s.number IS NOT NULL
    AND s.started_at >= sqlc.arg('started_from')
    AND s.started_at < sqlc.arg('started_before')
    AND w.deleted_at IS NULL
    AND w.unpublished_at IS NULL
ORDER BY s.started_at, s.id;

-- name: GetSlotForEpisodeGenerationByID :one
-- 作品ロック後に、現在も対象の放送枠だけを読み直す。所属作品はロックした作品に限る。
-- 放送枠もロックし、話数の計算から紐づけまでの間に回数・番組・公開状態を変更させない。
-- ほかのトランザクションがロックしている放送枠は待たずに飛ばし、次の実行で拾い直す。
-- Railsは放送枠を保存すると、放送枠の行を更新してから作品をtouchするため、作品を先に
-- ロックしたここで放送枠の行を待つと、互いの行を待ち合ってデッドロックする。
--
-- preceding_irregular_countは話数の計算で回数から引く不定期放送の数。この放送枠より回数が
-- 小さく、番組の下限以上のものだけを数える。後ろの不定期放送まで数えると、すでに紐づいた
-- 放送枠と同じ話数を求めてしまうため。
SELECT
    s.id,
    s.work_id,
    s.program_id,
    s.number,
    s.started_at,
    p.minimum_episode_generatable_number,
    (
        SELECT COUNT(*)
        FROM slots i
        WHERE i.program_id = s.program_id
            AND i.irregular
            AND i.deleted_at IS NULL
            AND i.unpublished_at IS NULL
            AND i.number >= p.minimum_episode_generatable_number
            AND i.number < s.number
    )::integer AS preceding_irregular_count
FROM slots s
JOIN programs p ON p.id = s.program_id
JOIN works w ON w.id = s.work_id
WHERE s.id = sqlc.arg('id')
    AND s.work_id = sqlc.arg('work_id')
    AND s.deleted_at IS NULL
    AND s.unpublished_at IS NULL
    AND s.episode_id IS NULL
    AND NOT s.rebroadcast
    AND NOT s.irregular
    AND s.number IS NOT NULL
    AND s.started_at >= sqlc.arg('started_from')
    AND s.started_at < sqlc.arg('started_before')
    AND w.deleted_at IS NULL
    AND w.unpublished_at IS NULL
FOR UPDATE OF s SKIP LOCKED;

-- name: LinkSlotEpisode :execrows
-- 照合に使った所属・番組・回数・開始時刻と対象条件を満たす放送枠だけを書き込む。
-- 0行なら呼び出し側は生成をロールバックする。Railsのupdate_column相当で、updated_atも
-- 作品のtouchも行わない。
UPDATE slots
SET episode_id = sqlc.arg('episode_id')
WHERE id = sqlc.arg('id')
    AND work_id = sqlc.arg('work_id')
    AND program_id = sqlc.arg('program_id')
    AND number = sqlc.arg('number')
    AND started_at = sqlc.arg('started_at')
    AND episode_id IS NULL
    AND deleted_at IS NULL
    AND unpublished_at IS NULL
    AND NOT rebroadcast
    AND NOT irregular;
