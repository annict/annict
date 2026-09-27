-- migrate:up

-- River (バックグラウンドジョブキュー) をRiver v0.40.0で追加されたスキーマ
-- マイグレーションバージョン7に追随させる。dbmateを単一の正本に保つため、本
-- マイグレーションは `river migrate-get --line main --version 7 --up/--down`
-- (go.modのRiverバージョンに固定したCLIで実行) が出力するSQL本体を取り込み、
-- 末尾のriver_migrationへのINSERTでバージョン7が適用済みであることを記録して、
-- RiverのGo API (rivermigrate) がスキーマを最新と認識できるようにする。これらの
-- テーブルは実行時にriverpgxv5経由でRiverが所有し、アプリケーションがsqlcを
-- 通じてクエリすることはない。
--
-- バージョン7は通知アウトボックスriver_notificationテーブル (+ インデックス2つ)
-- を追加し、未使用のriver_client / river_client_queueテーブルを削除し、
-- river_job.max_attempts (25) とriver_queue.updated_at (CURRENT_TIMESTAMP) に
-- カラムデフォルトを追加する。削除するテーブルはアプリケーションコードから参照されて
-- いないため、削除は安全。

CREATE TABLE river_notification (
    id bigserial PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    payload text NOT NULL,
    topic text NOT NULL,
    CONSTRAINT topic_length CHECK (length(topic) > 0 AND length(topic) < 128)
);

CREATE INDEX river_notification_created_at_idx ON river_notification (created_at);
CREATE INDEX river_notification_topic_id_idx ON river_notification (topic, id);

DROP TABLE river_client_queue;
DROP TABLE river_client;

ALTER TABLE river_job
    ALTER COLUMN max_attempts SET DEFAULT 25;

ALTER TABLE river_queue
    ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

INSERT INTO river_migration (line, version) VALUES ('main', 7);

-- migrate:down

CREATE UNLOGGED TABLE river_client (
    id text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}',
    paused_at timestamptz,
    updated_at timestamptz NOT NULL,
    CONSTRAINT name_length CHECK (char_length(id) > 0 AND char_length(id) < 128)
);

CREATE UNLOGGED TABLE river_client_queue (
    river_client_id text NOT NULL REFERENCES river_client (id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    max_workers bigint NOT NULL DEFAULT 0,
    metadata jsonb NOT NULL DEFAULT '{}',
    num_jobs_completed bigint NOT NULL DEFAULT 0,
    num_jobs_running bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (river_client_id, name),
    CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128),
    CONSTRAINT num_jobs_completed_zero_or_positive CHECK (num_jobs_completed >= 0),
    CONSTRAINT num_jobs_running_zero_or_positive CHECK (num_jobs_running >= 0)
);

ALTER TABLE river_job
    ALTER COLUMN max_attempts DROP DEFAULT;

ALTER TABLE river_queue
    ALTER COLUMN updated_at DROP DEFAULT;

DROP TABLE river_notification;

DELETE FROM river_migration WHERE line = 'main' AND version = 7;
