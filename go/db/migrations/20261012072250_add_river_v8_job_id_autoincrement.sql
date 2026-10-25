-- migrate:up

-- River (バックグラウンドジョブキュー) をRiver v0.48.0で追加されたスキーマ
-- マイグレーションバージョン8に追随させる。dbmateを単一の正本に保つため、本
-- マイグレーションは `river migrate-get --line main --version 8 --up/--down`
-- (go.modのRiverバージョンに固定したCLIで実行) が出力するSQL本体を取り込み、
-- 末尾のriver_migrationへのINSERTでバージョン8が適用済みであることを記録して、
-- RiverのGo API (rivermigrate) がスキーマを最新と認識できるようにする。
--
-- バージョン8はSQLiteドライバー向けにジョブIDの再利用を防ぐもので、PostgreSQLでは
-- シーケンスが既に再利用を防いでいるため、SQL本体はup / downともに何もしない。

-- River main migration 008 [up]
-- No-op. PostgreSQL sequences already prevent automatically generated job IDs
-- from being reused.
SELECT 1;

INSERT INTO river_migration (line, version) VALUES ('main', 8);

-- migrate:down

-- River main migration 008 [down]
-- No-op. PostgreSQL sequences already prevent automatically generated job IDs
-- from being reused.
SELECT 1;

DELETE FROM river_migration WHERE line = 'main' AND version = 8;
