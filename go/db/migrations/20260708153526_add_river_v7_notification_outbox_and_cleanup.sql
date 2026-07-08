-- migrate:up

-- Advance River (the background job queue) to schema migration version 7,
-- released in River v0.40.0. dbmate stays the single source of truth: this
-- migration vendors the clean SQL from `river migrate-get --line main --version
-- 7 --up/--down` (run with the CLI pinned to the go.mod River version), and the
-- closing INSERT into river_migration records that version 7 is applied so
-- River's Go API (rivermigrate) sees the schema as current. River owns these
-- tables at runtime via riverpgxv5; the application never queries them through
-- sqlc.
--
-- Version 7 adds the river_notification outbox table (with two indexes), drops
-- the unused river_client and river_client_queue tables, and adds column
-- defaults to river_job.max_attempts (25) and river_queue.updated_at
-- (CURRENT_TIMESTAMP). The dropped tables are never referenced by application
-- code, so removing them is safe.
--
-- [Ja] River (バックグラウンドジョブキュー) を River v0.40.0 で追加されたスキーマ
-- マイグレーションバージョン 7 に追随させる。dbmate を単一の正本に保つため、本
-- マイグレーションは `river migrate-get --line main --version 7 --up/--down`
-- (go.mod の River バージョンに固定した CLI で実行) が出力する clean SQL を取り込み、
-- 末尾の river_migration への INSERT でバージョン 7 が適用済みであることを記録して、
-- River の Go API (rivermigrate) がスキーマを最新と認識できるようにする。これらの
-- テーブルは実行時に riverpgxv5 経由で River が所有し、アプリケーションが sqlc を
-- 通じてクエリすることはない。
--
-- バージョン 7 は通知アウトボックス river_notification テーブル (+ インデックス 2 つ)
-- を追加し、未使用の river_client / river_client_queue テーブルを削除し、
-- river_job.max_attempts (25) と river_queue.updated_at (CURRENT_TIMESTAMP) に
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
