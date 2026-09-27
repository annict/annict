package testutil

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestSetupTx(t *testing.T) {
	t.Parallel()

	db, tx := SetupTx(t)

	if db == nil {
		t.Fatal("SetupTxが返した *sql.DBがnil")
	}
	if tx == nil {
		t.Fatal("SetupTxが返した *sql.Txがnil")
	}

	// トランザクション内でクエリが実行できることを確認
	var got int
	if err := tx.QueryRowContext(context.Background(), "SELECT 1").Scan(&got); err != nil {
		t.Fatalf("トランザクション内のクエリに失敗: %v", err)
	}
	if got != 1 {
		t.Errorf("SELECT 1の結果 = %d、期待値 = 1", got)
	}
}

// TestSetupTx_AutoRollbackはSetupTxが登録するt.Cleanupにより
// テスト終了時にトランザクションが必ずロールバック (クローズ) されることを検証します。
// 後続フェーズで多数のテストがSetupTxに置換されるため、責務の動作保証を強めておく狙いです。
func TestSetupTx_AutoRollback(t *testing.T) {
	t.Parallel()

	var capturedTx *sql.Tx

	// サブテスト終了時点でSetupTxのCleanupが走り、capturedTxはロールバック済みになる
	t.Run("内側", func(t *testing.T) {
		_, tx := SetupTx(t)
		capturedTx = tx

		// このスコープ内ではtxはまだ生きている
		var got int
		if err := tx.QueryRowContext(context.Background(), "SELECT 1").Scan(&got); err != nil {
			t.Fatalf("ロールバック前のクエリで失敗: %v", err)
		}
	})

	// サブテスト終了後はロールバック済みのため、Commitはsql.ErrTxDoneを返す
	if err := capturedTx.Commit(); !errors.Is(err, sql.ErrTxDone) {
		t.Errorf("ロールバック後のCommit()のエラー = %v、期待値 = sql.ErrTxDone", err)
	}
}
