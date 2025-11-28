// Package testutil はテスト用ヘルパー関数を提供します
package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/annict/annict/internal/query"
	_ "github.com/lib/pq"
)

var (
	testDB     *sql.DB
	testDBOnce sync.Once
)

// SetupTestDB はテスト用のデータベース接続を初期化し、トランザクションを開始します
// 各テスト終了時には自動的にロールバックされるため、データベースの状態がクリーンに保たれます
func SetupTestDB(t *testing.T) (*sql.DB, *sql.Tx) {
	t.Helper()

	// テスト用データベース接続の初期化
	// すべての環境でGoプロセス起動時には既に環境変数がセット済みです：
	// - ローカル開発/テスト: op run --env-file=".env" が処理済み
	// - CI環境: GitHub Actionsが設定済み
	testDBOnce.Do(func() {
		// テスト用データベースの接続情報
		// Dev Container内: docker-compose.ymlのpostgresqlサービス（postgresql:5432）
		// CI: GitHub Actionsのpostgresqlサービス（localhost:5432）
		dsn := getEnv("DATABASE_URL", "postgres://postgres@postgresql:5432/annict_test?sslmode=disable")

		// データベース接続の確立
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			panic(fmt.Sprintf("テスト用データベースへの接続に失敗しました: %v", err))
		}

		// 接続プールの設定
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)

		// 接続確認
		if err := db.Ping(); err != nil {
			panic(fmt.Sprintf("テスト用データベースへのping失敗: %v", err))
		}

		testDB = db
	})

	// トランザクションの開始
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatalf("トランザクションの開始に失敗しました: %v", err)
	}

	// テスト終了時にロールバック
	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Errorf("トランザクションのロールバックに失敗しました: %v", err)
		}
	})

	return testDB, tx
}

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返します
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TruncateTables は指定されたテーブルのデータを削除します（テスト間でのクリーンアップ用）
func TruncateTables(tx *sql.Tx, tables ...string) error {
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("テーブル %s のTRUNCATEに失敗: %w", table, err)
		}
	}
	return nil
}

// NewQueriesWithTx はトランザクションを使用するsqlc Queriesを作成します
func NewQueriesWithTx(db *sql.DB, tx *sql.Tx) *query.Queries {
	return query.New(db).WithTx(tx)
}

// GetTestDB はテスト用のデータベース接続を返します（並列実行対応）
func GetTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// テスト用データベース接続の初期化
	// すべての環境でGoプロセス起動時には既に環境変数がセット済みです：
	// - ローカル開発/テスト: op run --env-file=".env" が処理済み
	// - CI環境: GitHub Actionsが設定済み
	testDBOnce.Do(func() {
		// テスト用データベースの接続情報
		// Dev Container内: docker-compose.ymlのpostgresqlサービス（postgresql:5432）
		// CI: GitHub Actionsのpostgresqlサービス（localhost:5432）
		dsn := getEnv("DATABASE_URL", "postgres://postgres@postgresql:5432/annict_test?sslmode=disable")

		// データベース接続の確立
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			panic(fmt.Sprintf("テスト用データベースへの接続に失敗しました: %v", err))
		}

		// 接続プールの設定
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)

		// 接続確認
		if err := db.Ping(); err != nil {
			panic(fmt.Sprintf("テスト用データベースへのping失敗: %v", err))
		}

		testDB = db
	})

	return testDB
}
