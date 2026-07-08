package worker

import (
	"testing"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

// TestAppliedRiverMigrationVersionMatchesLibrary guards against River's schema
// drifting ahead of this project's dbmate migrations. rivermigrate.AllVersions()
// reads only the migration versions embedded in the linked River library (no
// database I/O), so its maximum reflects the newest version the library knows
// about. When a River upgrade (e.g. a dependabot bump) introduces a version
// beyond appliedRiverMigrationVersion, this test fails and signals that a dbmate
// migration must be added to follow River's new schema. The reproducible
// follow-up procedure lives in the create_river_migration_tracking migration
// header.
//
// [Ja] 本テストは River のスキーマが本プロジェクトの dbmate マイグレーションより
// 先行して乖離するのを防ぐ。rivermigrate.AllVersions() はリンク済みの River
// ライブラリに埋め込まれたマイグレーションバージョンを読むだけで (データベース
// I/O なし)、その最大値はライブラリが認識する最新バージョンを表す。River の
// アップグレード (dependabot による bump など) が appliedRiverMigrationVersion を
// 超えるバージョンを導入すると本テストは失敗し、River の新スキーマに追随する
// dbmate マイグレーションの追加が必要であることを知らせる。再現可能な追随手順は
// create_river_migration_tracking マイグレーションのヘッダーにある。
func TestAppliedRiverMigrationVersionMatchesLibrary(t *testing.T) {
	t.Parallel()

	// A nil pool is safe here: AllVersions() only reads embedded migration files
	// and never touches the database.
	//
	// [Ja] ここでは nil プールで安全: AllVersions() は埋め込みのマイグレーション
	// ファイルを読むだけで、データベースには一切アクセスしない。
	migrator, err := rivermigrate.New(riverpgxv5.New(nil), nil)
	if err != nil {
		t.Fatalf("rivermigrate.New() でエラー: %v", err)
	}

	versions := migrator.AllVersions()
	if len(versions) == 0 {
		t.Fatal("AllVersions() が空のスライスを返した")
	}

	// AllVersions() is sorted ascending by version, so the last element is the
	// newest version the linked River library knows about.
	//
	// [Ja] AllVersions() はバージョン昇順にソートされているため、末尾の要素が
	// リンク済みの River ライブラリが認識する最新バージョン。
	latest := versions[len(versions)-1].Version

	if latest != appliedRiverMigrationVersion {
		t.Errorf(
			"River ライブラリの最新マイグレーション version (%d) が適用済み version 定数 appliedRiverMigrationVersion (%d) と一致しません。"+
				"River が新しいマイグレーションを導入した可能性があります。"+
				"`river migrate-get --line main --version %d --up/--down` で SQL を生成して dbmate マイグレーションを追加し、"+
				"appliedRiverMigrationVersion を %d に更新してください。",
			latest, appliedRiverMigrationVersion, latest, latest,
		)
	}
}
