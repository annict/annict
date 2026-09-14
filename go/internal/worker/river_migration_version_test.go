package worker

import (
	"testing"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

// 本テストはRiverのスキーマが本プロジェクトのdbmateマイグレーションより
// 先行して乖離するのを防ぐ。rivermigrate.AllVersions() はリンク済みのRiver
// ライブラリに埋め込まれたマイグレーションバージョンを読むだけで (データベース
// I/Oなし)、その最大値はライブラリが認識する最新バージョンを表す。Riverの
// アップグレード (dependabotによるbumpなど) がappliedRiverMigrationVersionを
// 超えるバージョンを導入すると本テストは失敗し、Riverの新スキーマに追随する
// dbmateマイグレーションの追加が必要であることを知らせる。再現可能な追随手順は
// create_river_migration_trackingマイグレーションのヘッダーにある。
func TestAppliedRiverMigrationVersionMatchesLibrary(t *testing.T) {
	t.Parallel()

	// ここではnilプールで安全: AllVersions() は埋め込みのマイグレーション
	// ファイルを読むだけで、データベースには一切アクセスしない。
	migrator, err := rivermigrate.New(riverpgxv5.New(nil), nil)
	if err != nil {
		t.Fatalf("rivermigrate.New()でエラー: %v", err)
	}

	versions := migrator.AllVersions()
	if len(versions) == 0 {
		t.Fatal("AllVersions()が空のスライスを返した")
	}

	// AllVersions() はバージョン昇順にソートされているため、末尾の要素が
	// リンク済みのRiverライブラリが認識する最新バージョン。
	latest := versions[len(versions)-1].Version

	if latest != appliedRiverMigrationVersion {
		t.Errorf(
			"Riverライブラリの最新マイグレーションversion (%d) が適用済みversion定数appliedRiverMigrationVersion (%d) と一致しません。"+
				"Riverが新しいマイグレーションを導入した可能性があります。"+
				"`river migrate-get --line main --version %d --up/--down`でSQLを生成してdbmateマイグレーションを追加し、"+
				"appliedRiverMigrationVersionを%dに更新してください。",
			latest, appliedRiverMigrationVersion, latest, latest,
		)
	}
}
