-- migrate:up

-- Introduce river_migration, the tracking table River (the background job
-- queue) uses to record which of its own schema migrations are applied. This
-- project vendors River's DDL as dbmate migrations and keeps a single dumped
-- db/schema.sql as the source of truth instead of running River's own migrator,
-- so the tracking table was never created. River's tables themselves already
-- exist from the earlier migration 20251021153929_add_river_tables.sql, which
-- reproduced River v0.39.0's schema up to version 6, so this migration is
-- limited to adding river_migration and seeding it with versions 1..6 to record
-- that those six main-line migrations are applied. Seeding lets River's Go API
-- (rivermigrate) recognize the applied versions so the schema can be verified
-- and advanced. River owns these tables at runtime via riverpgxv5; the
-- application never queries them through sqlc.
--
-- When River is bumped and a new migration version N appears, keep dbmate as
-- the source of truth: generate clean SQL with `river migrate-get --line main
-- --version N --up/--down` (pinned via `go run .../cmd/river@vX.Y.Z`, excluding
-- version 1), add it as a new dbmate migration, and append `INSERT INTO
-- river_migration (line, version) VALUES ('main', N);` (DELETE on down).
--
-- [Ja] River (バックグラウンドジョブキュー) が自身のどのスキーママイグレーション
-- を適用済みかを記録する追跡テーブル river_migration を導入する。本プロジェクトは
-- River 独自のマイグレータを走らせず、River の DDL を dbmate マイグレーションとして
-- 取り込み、ダンプ済みの単一 db/schema.sql を正本としているため、この追跡テーブルは
-- これまで作られていなかった。River のテーブル自体は River v0.39.0 のバージョン 6
-- までのスキーマを再現した先行マイグレーション 20251021153929_add_river_tables.sql で
-- 既に存在するので、本マイグレーションは river_migration の追加に閉じ、それら 6 つの
-- main ラインマイグレーションが適用済みであることを記録するために version 1..6 を
-- seed する。seed により River の Go API (rivermigrate) が適用済みバージョンを認識
-- でき、スキーマの検証・追随が可能になる。これらのテーブルは実行時に riverpgxv5 経由
-- で River が所有し、アプリケーションが sqlc を通じてクエリすることはない。
--
-- River を bump して新しいマイグレーションバージョン N が増えたときは、dbmate を正本
-- に保つ: `river migrate-get --line main --version N --up/--down` (`go run
-- .../cmd/river@vX.Y.Z` でバージョン固定、version 1 は除外) で clean SQL を生成し、
-- 新しい dbmate マイグレーションとして追加し、末尾に `INSERT INTO river_migration
-- (line, version) VALUES ('main', N);` を追記する (down では DELETE)。

CREATE TABLE river_migration(
    line TEXT NOT NULL,
    version bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    CONSTRAINT line_length CHECK (char_length(line) > 0 AND char_length(line) < 128),
    CONSTRAINT version_gte_1 CHECK (version >= 1),
    PRIMARY KEY (line, version)
);

INSERT INTO river_migration (line, version) VALUES
    ('main', 1),
    ('main', 2),
    ('main', 3),
    ('main', 4),
    ('main', 5),
    ('main', 6);

-- migrate:down

DROP TABLE river_migration;
