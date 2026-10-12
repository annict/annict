-- migrate:up

-- River (バックグラウンドジョブキュー) をRiver v0.49.0で追加されたスキーマ
-- マイグレーションバージョン8に追随させる。`river migrate-get --line main
-- --version 8 --up/--down` (go.modのRiverバージョンに固定したCLIで実行) が出力する
-- SQL本体は、PostgreSQLではup / downともに `SELECT 1;` だけのno-opである。
-- バージョン8はSQLiteドライバー向けにジョブIDの再利用を防ぐもので、PostgreSQLでは
-- シーケンスが既に再利用を防いでいるためスキーマの変更は無い。そのため本
-- マイグレーションはriver_migrationへのINSERTでバージョン8が適用済みであることを
-- 記録するだけにとどめ、RiverのGo API (rivermigrate) がスキーマを最新と認識できる
-- ようにする。

INSERT INTO river_migration (line, version) VALUES ('main', 8);

-- migrate:down

DELETE FROM river_migration WHERE line = 'main' AND version = 8;
