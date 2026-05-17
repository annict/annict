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
		t.Fatal("SetupTx が返した *sql.DB が nil")
	}
	if tx == nil {
		t.Fatal("SetupTx が返した *sql.Tx が nil")
	}

	// トランザクション内でクエリが実行できることを確認
	var got int
	if err := tx.QueryRowContext(context.Background(), "SELECT 1").Scan(&got); err != nil {
		t.Fatalf("トランザクション内のクエリに失敗: %v", err)
	}
	if got != 1 {
		t.Errorf("SELECT 1 の結果が想定外: got=%d, want=1", got)
	}
}

// TestSetupTx_AutoRollback は SetupTx が登録する t.Cleanup により
// テスト終了時にトランザクションが必ずロールバック（クローズ）されることを検証します。
// 後続フェーズで多数のテストが SetupTx に置換されるため、責務の動作保証を強めておく狙いです。
func TestSetupTx_AutoRollback(t *testing.T) {
	t.Parallel()

	var capturedTx *sql.Tx

	// サブテスト終了時点で SetupTx の Cleanup が走り、capturedTx はロールバック済みになる
	t.Run("inner", func(t *testing.T) {
		_, tx := SetupTx(t)
		capturedTx = tx

		// このスコープ内では tx はまだ生きている
		var got int
		if err := tx.QueryRowContext(context.Background(), "SELECT 1").Scan(&got); err != nil {
			t.Fatalf("ロールバック前のクエリで失敗: %v", err)
		}
	})

	// サブテスト終了後はロールバック済みのため、Commit は sql.ErrTxDone を返す
	if err := capturedTx.Commit(); !errors.Is(err, sql.ErrTxDone) {
		t.Errorf("ロールバック後の Commit は sql.ErrTxDone を期待: got=%v", err)
	}
}
