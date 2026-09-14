-- migrate:up

-- River (バックグラウンドジョブキュー) が自身のどのスキーママイグレーション
-- を適用済みかを記録する追跡テーブルriver_migrationを導入する。本プロジェクトは
-- River独自のマイグレータを走らせず、RiverのDDLをdbmateマイグレーションとして
-- 取り込み、ダンプ済みの単一db/schema.sqlを正本としているため、この追跡テーブルは
-- これまで作られていなかった。Riverのテーブル自体はRiver v0.39.0のバージョン6
-- までのスキーマを再現した先行マイグレーション20251021153929_add_river_tables.sqlで
-- 既に存在するので、本マイグレーションはriver_migrationの追加に閉じ、それら6つの
-- mainラインマイグレーションが適用済みであることを記録するために、version列が
-- 1から6の行を初期データとして追加する。これによりRiverのGo API (rivermigrate) が適用済みバージョンを認識
-- でき、スキーマの検証・追随が可能になる。これらのテーブルは実行時にriverpgxv5経由
-- でRiverが所有し、アプリケーションがsqlcを通じてクエリすることはない。
--
-- Riverのバージョンを上げて新しいマイグレーションバージョンNが増えたときは、dbmateを正本
-- に保つ: `river migrate-get --line main --version N --up/--down` (`go run
-- .../cmd/river@vX.Y.Z` でバージョン固定、バージョン1は除外) でSQL本体を生成し、
-- 新しいdbmateマイグレーションとして追加し、末尾に `INSERT INTO river_migration
-- (line, version) VALUES ('main', N);` を追記する (downではDELETE)。

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
