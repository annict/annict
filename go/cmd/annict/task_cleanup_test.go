package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

// cleanupTaskBodiesは、登録済みのクリーンアップタスク名と、そのクリーンアップUseCase
// を組み立てて呼ぶタスク本体を対応づける。期待する本体を登録名で引ける形にすることで、
// 本番レジストリとDB配線を独立して検証できる。
var cleanupTaskBodies = map[string]taskBody{
	"cleanup-expired-tokens":        cleanupExpiredTokens,
	"cleanup-expired-sign-in-codes": cleanupExpiredSignInCodes,
	"cleanup-expired-sessions":      cleanupExpiredSessions,
}

// TestCleanupTaskBodies_Registeredは、各タスク名がそれぞれのタスク本体に解決すること
// を確認する。もう一方のクリーンアップの本体を登録してしまうと、運用者が指定していない
// テーブルの行が削除されるが、本パッケージの他のテストではそれを検出できない。
func TestCleanupTaskBodies_Registered(t *testing.T) {
	t.Parallel()

	for name, want := range cleanupTaskBodies {
		task, ok := tasks[name]
		if !ok {
			t.Errorf("タスク%qが登録されていない", name)
			continue
		}
		if task.body == nil {
			t.Errorf("タスク%qにbodyが無い", name)
			continue
		}
		if got, want := reflect.ValueOf(task.body).Pointer(), reflect.ValueOf(want).Pointer(); got != want {
			t.Errorf("タスク%qが誤ったbodyで登録されている", name)
		}
	}
}

// TestCleanupTaskBodies_Wiringは各タスク本体をテスト用データベースに対して実行する。
// 本体はUseCaseを組み立てて呼ぶだけであるため、CLI側で確かめるべきは、配線が実スキーマの
// 受け付けるSQLまで解決できることである。何が削除されるかはUseCase側の関心事。
//
// 設定をnilにしているのは、これらの本体が設定を読まないため。必要なのはsqlcのクエリで
// あり、検証対象の配線が解決するのもそれである。
//
// クエリはテスト用トランザクション上に組み立て、DELETEがロールバックされてテスト用
// データベースを共有する他パッケージから見えないようにする。
func TestCleanupTaskBodies_Wiring(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	for name, body := range cleanupTaskBodies {
		t.Run(name, func(t *testing.T) {
			if err := body(context.Background(), nil, db, testutil.NewQueriesWithTx(db, tx)); err != nil {
				t.Errorf("%s()のエラー = %v", name, err)
			}
		})
	}
}

// TestCleanupTaskBodies_Errorは、UseCaseの失敗が握り潰されずerrorとして返ることを
// 確認する。mainはそのerrorを非ゼロの終了コードに変換するため。失敗は、クエリが乗って
// いるトランザクションを先に終了させることで起こす。
func TestCleanupTaskBodies_Error(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	if err := tx.Rollback(); err != nil {
		t.Fatalf("トランザクションのロールバックに失敗しました: %v", err)
	}

	for name, body := range cleanupTaskBodies {
		t.Run(name, func(t *testing.T) {
			if err := body(context.Background(), nil, db, testutil.NewQueriesWithTx(db, tx)); err == nil {
				t.Errorf("%s()のエラー = nil、期待値 = エラーあり", name)
			}
		})
	}
}
