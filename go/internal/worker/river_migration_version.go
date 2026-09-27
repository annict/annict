package worker

// appliedRiverMigrationVersionは、本プロジェクトのdbmateマイグレーションが
// 適用し追跡テーブルriver_migrationに記録済みの、River (バックグラウンドジョブ
// キュー) スキーママイグレーションの最大バージョン。本パッケージのドリフト検知
// テストの基準値であり、RiverのGoライブラリがこの番号を超えるマイグレーション
// バージョンを提供すると、そのテストが失敗してdbmate側の追随を促し、スキーマが
// リンク済みライブラリから静かに遅れることを防ぐ。
//
// 更新するときは、create_river_migration_trackingマイグレーションのヘッダーに記した
// 再現手順に従う: `river migrate-get --line main --version N --up/--down` (go.modの
// Riverバージョンに固定したCLI、version 1は除外) でclean SQLを生成し、末尾に
// `INSERT INTO river_migration (line, version) VALUES ('main', N);` を追記する
// (downではDELETE) 新しいdbmateマイグレーションを追加し、この定数をNに更新する。
const appliedRiverMigrationVersion = 7
