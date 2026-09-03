package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

// cleanupTaskBodies pairs each registered cleanup task name with the body that wires and
// calls its cleanup usecase. Holding the expected body under the registered name lets the
// tests check the production registry and the database wiring independently.
//
// [Ja] cleanupTaskBodies は、登録済みのクリーンアップタスク名と、そのクリーンアップ UseCase
// を組み立てて呼ぶタスク本体を対応づける。期待する本体を登録名で引ける形にすることで、
// 本番レジストリと DB 配線を独立して検証できる。
var cleanupTaskBodies = map[string]taskBody{
	"cleanup-expired-tokens":        cleanupExpiredTokens,
	"cleanup-expired-sign-in-codes": cleanupExpiredSignInCodes,
	"cleanup-expired-sessions":      cleanupExpiredSessions,
}

// TestCleanupTaskBodies_Registered checks that each task name resolves to its own body.
// Registering a name with the other cleanup's body would delete rows from a table the
// operator did not ask about, and nothing else in the package would catch it.
//
// [Ja] TestCleanupTaskBodies_Registered は、各タスク名がそれぞれのタスク本体に解決すること
// を確認する。もう一方のクリーンアップの本体を登録してしまうと、運用者が指定していない
// テーブルの行が削除されるが、本パッケージの他のテストではそれを検出できない。
func TestCleanupTaskBodies_Registered(t *testing.T) {
	t.Parallel()

	for name, want := range cleanupTaskBodies {
		task, ok := tasks[name]
		if !ok {
			t.Errorf("task %q is not registered", name)
			continue
		}
		if task.body == nil {
			t.Errorf("task %q has no body", name)
			continue
		}
		if got, want := reflect.ValueOf(task.body).Pointer(), reflect.ValueOf(want).Pointer(); got != want {
			t.Errorf("task %q is registered with the wrong body", name)
		}
	}
}

// TestCleanupTaskBodies_Wiring runs each body against the test database. The bodies do
// nothing but wire a usecase and call it, so what has to be proven on the CLI side is
// that the wiring resolves all the way down to SQL the real schema accepts. What gets
// deleted is the usecase's own concern.
//
// The configuration is nil because these bodies read none: what they need is the sqlc
// queries, which is what the wiring under test resolves.
//
// The queries are built on the test transaction so that the DELETE rolls back and stays
// invisible to the other packages sharing the test database.
//
// [Ja] TestCleanupTaskBodies_Wiring は各タスク本体をテスト用データベースに対して実行する。
// 本体は UseCase を組み立てて呼ぶだけであるため、CLI 側で確かめるべきは、配線が実スキーマの
// 受け付ける SQL まで解決できることである。何が削除されるかは UseCase 側の関心事。
//
// 設定を nil にしているのは、これらの本体が設定を読まないため。必要なのは sqlc のクエリで
// あり、検証対象の配線が解決するのもそれである。
//
// クエリはテスト用トランザクション上に組み立て、DELETE がロールバックされてテスト用
// データベースを共有する他パッケージから見えないようにする。
func TestCleanupTaskBodies_Wiring(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)

	for name, body := range cleanupTaskBodies {
		t.Run(name, func(t *testing.T) {
			if err := body(context.Background(), nil, db, testutil.NewQueriesWithTx(db, tx)); err != nil {
				t.Errorf("%s() error = %v", name, err)
			}
		})
	}
}

// TestCleanupTaskBodies_Error checks that a failing usecase comes back as an error
// instead of being swallowed, since that error is what main turns into a non-zero exit
// code. The failure is produced by finishing the transaction the queries run on.
//
// [Ja] TestCleanupTaskBodies_Error は、UseCase の失敗が握り潰されず error として返ることを
// 確認する。main はその error を非ゼロの終了コードに変換するため。失敗は、クエリが乗って
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
				t.Errorf("%s() error = nil, want error", name)
			}
		})
	}
}
