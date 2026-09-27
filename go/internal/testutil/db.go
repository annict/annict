// Package testutilはテスト用ヘルパー関数を提供します
package testutil

import (
	"cmp"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/query"
)

var (
	testDB     *sql.DB
	testDBOnce sync.Once
)

// SetupTestMainはテストパッケージごとのTestMainで呼び出すヘルパー関数です
// bcryptコストの低減とDB接続プールの初期化をパッケージ内で1度だけ行ってからm.Run() を実行します
// 戻り値はos.Exitに渡すための終了コードです
func SetupTestMain(m *testing.M) int {
	initTestDB()
	return m.Run()
}

// SetupTxはテスト用のトランザクションを開始します
// テスト終了時には自動的にロールバックされるため、データベースの状態がクリーンに保たれます
func SetupTx(t *testing.T) (*sql.DB, *sql.Tx) {
	t.Helper()

	initTestDB()

	tx, err := testDB.Begin()
	if err != nil {
		t.Fatalf("トランザクションの開始に失敗しました: %v", err)
	}

	t.Cleanup(func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			t.Errorf("トランザクションのロールバックに失敗しました: %v", err)
		}
	})

	return testDB, tx
}

// GetTestDBはテスト用のデータベース接続を返します
// SetupTestMainで接続が初期化されていることを前提とし、未初期化の場合は念のため初期化します
func GetTestDB() *sql.DB {
	initTestDB()
	return testDB
}

// initTestDBはテスト用DB接続プールの初期化をsync.Onceにより1度だけ実行します
// SetupTestMain / SetupTx / GetTestDBのいずれから呼ばれても同じ接続を共有します
func initTestDB() {
	testDBOnce.Do(func() {
		// テスト用にbcryptコストを下げる
		auth.SetBcryptCostForTest(bcrypt.MinCost)

		// DATABASE_URLはop run / GitHub Actionsが事前にセット済み。
		// 未設定時はDev Containerのpostgresqlサービスをデフォルトとする。
		dsn := cmp.Or(os.Getenv("DATABASE_URL"), "postgres://postgres@postgresql:5432/annict_test?sslmode=disable")

		db, err := sql.Open("postgres", dsn)
		if err != nil {
			panic(fmt.Sprintf("テスト用データベースへの接続に失敗しました: %v", err))
		}

		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)

		if err := db.Ping(); err != nil {
			panic(fmt.Sprintf("テスト用データベースへのping失敗: %v", err))
		}

		testDB = db
	})
}

// TruncateTablesは指定されたテーブルのデータを削除します (テスト間でのクリーンアップ用)
func TruncateTables(tx *sql.Tx, tables ...string) error {
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("テーブル %sのTRUNCATEに失敗: %w", table, err)
		}
	}
	return nil
}

// NewQueriesWithTxはトランザクションを使用するsqlc Queriesを作成します
func NewQueriesWithTx(db *sql.DB, tx *sql.Tx) *query.Queries {
	return query.New(db).WithTx(tx)
}
