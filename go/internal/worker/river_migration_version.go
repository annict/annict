package worker

// appliedRiverMigrationVersion is the highest River (the background job queue)
// schema migration version that this project's dbmate migrations have applied
// and recorded in the river_migration tracking table. It anchors the
// drift-detection test in this package: when River's Go library ships a
// migration version beyond this number, that test fails and prompts a dbmate
// follow-up so the schema never silently lags behind the linked library.
//
// To advance it, follow the reproducible procedure documented in the
// create_river_migration_tracking migration header: generate clean SQL with
// `river migrate-get --line main --version N --up/--down` (CLI pinned to the
// go.mod River version, version 1 excluded), add it as a new dbmate migration
// that appends `INSERT INTO river_migration (line, version) VALUES ('main', N);`
// (DELETE on down), then bump this constant to N.
//
// [Ja] appliedRiverMigrationVersion は、本プロジェクトの dbmate マイグレーションが
// 適用し追跡テーブル river_migration に記録済みの、River (バックグラウンドジョブ
// キュー) スキーママイグレーションの最大バージョン。本パッケージのドリフト検知
// テストの基準値であり、River の Go ライブラリがこの番号を超えるマイグレーション
// バージョンを提供すると、そのテストが失敗して dbmate 側の追随を促し、スキーマが
// リンク済みライブラリから静かに遅れることを防ぐ。
//
// 更新するときは、create_river_migration_tracking マイグレーションのヘッダーに記した
// 再現手順に従う: `river migrate-get --line main --version N --up/--down` (go.mod の
// River バージョンに固定した CLI、version 1 は除外) で clean SQL を生成し、末尾に
// `INSERT INTO river_migration (line, version) VALUES ('main', N);` を追記する
// (down では DELETE) 新しい dbmate マイグレーションを追加し、この定数を N に更新する。
const appliedRiverMigrationVersion = 7
