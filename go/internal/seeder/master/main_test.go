package master

import (
	"cmp"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/annict/annict/go/internal/testutil"
)

// schemaPathは適用するスキーマの正本。Makefileのdb-setup-testと同じファイルを使い、
// 列定義の取得先が本物のスキーマから外れないようにする。
const schemaPath = "../../../db/schema.sql"

var (
	// baseDBURLはテスト用データベースの接続URL。使い捨てDBの接続先はDB名だけを差し替えて作る。
	baseDBURL *url.URL

	// templateDBNameはスキーマを適用済みの複製元データベース。
	templateDBName string
)

// TestMainはスキーマを適用したテンプレート用データベースを1つ用意する。
//
// このパッケージのDBテストは、共有テストDBではなくテストごとの使い捨てDBで動く。setvalが
// ロールバックで戻らないため、テーブルだけでなくシーケンスも隔離する必要があるため。
// 使い捨てDBをテンプレートから複製すれば、8,000行を超えるスキーマの適用はパッケージ全体で
// 1回で済む。
func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

// runTestsはテンプレート用データベースの後始末をdeferで書けるよう、TestMainから分けている。
// os.Exitはdeferを実行しない。
func runTests(m *testing.M) int {
	dsn := cmp.Or(os.Getenv("DATABASE_URL"), "postgres://postgres@postgresql:5432/annict_test?sslmode=disable")

	parsed, err := url.Parse(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "テストDBの接続URLを解析できない: %v\n", err)

		return 1
	}
	baseDBURL = parsed
	templateDBName = newDBName("template_")

	admin := testutil.GetTestDB()
	if _, err := admin.Exec("CREATE DATABASE " + pq.QuoteIdentifier(templateDBName)); err != nil {
		fmt.Fprintf(os.Stderr, "テンプレート用DBの作成に失敗した: %v\n", err)

		return 1
	}
	defer func() {
		if _, err := admin.Exec("DROP DATABASE " + pq.QuoteIdentifier(templateDBName)); err != nil {
			fmt.Fprintf(os.Stderr, "テンプレート用DBの削除に失敗した: %v\n", err)
		}
	}()

	if err := applySchema(dbURLFor(templateDBName)); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)

		return 1
	}

	return m.Run()
}

// newDBNameは使い捨てのデータベース名を作る。同時に走るテストとの衝突を避けるためUUIDを
// 付ける。PostgreSQLの識別子は63バイトまでで、この形は最長でも56バイトに収まる。
func newDBName(prefix string) string {
	return "annict_test_master_" + prefix + strings.ReplaceAll(uuid.NewString(), "-", "")
}

// dbURLForはテスト用データベースの接続URLのDB名だけを差し替えて返す。
func dbURLFor(name string) string {
	dbURL := *baseDBURL
	dbURL.Path = "/" + name

	return dbURL.String()
}

// applySchemaはスキーマの正本をpsqlで流す。dbmateはCIのランナーに無いため、Makefileの
// db-setup-testと同じくpsqlを使う。
func applySchema(dsn string) error {
	cmd := exec.Command("psql", "--no-psqlrc", "--dbname", dsn, "--set", "ON_ERROR_STOP=1", "--file", schemaPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("テンプレート用DBへのスキーマ適用に失敗した: %w\n%s", err, output)
	}

	return nil
}
