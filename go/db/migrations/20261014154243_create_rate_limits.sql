-- migrate:up

-- レート制限のカウンターを固定窓ごとに保持するテーブル。Go版の限定器 (internal/ratelimit) と
-- Rails版のrack-attackが共有し、キーの名前空間で数える対象を分ける。
--
-- 全リクエストで1回upsertするため、書き込みの負荷を抑える設定にしている。
-- - UNLOGGED: WALを書かない。クラッシュ時に中身が消えるが、カウンターは失われてよい
-- - fillfactor: 同じページに更新後の行を置く余地を残し、HOT更新にする
-- - countにはインデックスを張らない (張るとHOT更新にならない)
-- - 期限切れの窓の削除は定期ジョブで行う。テーブルを小さく保つため、window_startにも
--   インデックスを張らず、削除は全件の走査で済ませる
CREATE UNLOGGED TABLE rate_limits (
    key character varying NOT NULL,
    window_start timestamp with time zone NOT NULL,
    count integer NOT NULL DEFAULT 1,
    PRIMARY KEY (key, window_start)
) WITH (fillfactor = 50);

-- migrate:down

DROP TABLE rate_limits;
