-- name: ListSlotsForEpisodeGeneration :many
-- 自動生成の対象はRailsのepisode:generateと同じ条件 (未削除・公開中で、エピソードが紐づいて
-- おらず、再放送でも不定期放送でもなく、番組と回数を持つ) の放送枠に、開始日時の範囲と作品が
-- 未削除・公開中であることを加えたもの。
--
-- preceding_irregular_countは話数の計算で回数から引く不定期放送の数。この放送枠より回数が
-- 小さいものだけを数える。後ろの不定期放送まで数えると、すでに紐づいた放送枠と同じ話数を
-- 求めてしまうため。
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
WHERE s.deleted_at IS NULL
    AND s.unpublished_at IS NULL
    AND s.episode_id IS NULL
    AND NOT s.rebroadcast
    AND NOT s.irregular
    AND s.number IS NOT NULL
    AND s.started_at >= sqlc.arg('started_from')
    AND s.started_at < sqlc.arg('started_before')
    AND w.deleted_at IS NULL
    AND w.unpublished_at IS NULL
ORDER BY s.started_at, s.id;

-- name: LinkSlotEpisode :execrows
-- 一覧を読んだ後に、編集者や別の実行が同じ放送枠へエピソードを紐づけていることがある。
-- 紐づいていない間だけ書き、0行なら呼び出し側はその放送枠を飛ばす。Railsの自動生成と同じく
-- update_column相当で、updated_atも作品のtouchも行わない。
UPDATE slots
SET episode_id = sqlc.arg('episode_id')
WHERE id = sqlc.arg('id')
    AND episode_id IS NULL
    AND deleted_at IS NULL;
