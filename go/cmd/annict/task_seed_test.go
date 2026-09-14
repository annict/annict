package main

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
)

// TestSeedTask_Registeredはseedというタスク名が環境ガードと本体に解決することを確認
// する。運用者が打ち込むのはタスク名だけであるため、登録はDB接続前の安全確認と要求された
// 処理の両方を保たなければならない。
func TestSeedTask_Registered(t *testing.T) {
	t.Parallel()

	task, ok := tasks["seed"]
	if !ok {
		t.Fatal(`タスク"seed"が登録されていない`)
	}
	if task.body == nil {
		t.Fatal(`タスク"seed"にbodyが無い`)
	}
	if task.guard == nil {
		t.Fatal(`タスク"seed"にguardが無い`)
	}
	if got, want := reflect.ValueOf(task.guard).Pointer(), reflect.ValueOf(guardSeed).Pointer(); got != want {
		t.Error(`タスク"seed"が誤ったguardで登録されている`)
	}
	if got, want := reflect.ValueOf(task.body).Pointer(), reflect.ValueOf(seed).Pointer(); got != want {
		t.Error(`タスク"seed"が誤ったbodyで登録されている`)
	}
}

// TestSeedTask_RejectsNonSeedableEnvBeforeDBConnectionは、意図的に接続できない
// データベースを設定して登録済みタスクを実行する。本番環境はタスクガードが先に拒否しなければ
// ならない。sql.OpenやPingContextまで到達すると環境エラーが接続エラーに置き換わり、本テスト
// は失敗する。
//
// config.Loadはプロセスの環境変数を読み、t.Setenvがテスト中にその値を変えるため、本テストは
// 並列実行できない。
func TestSeedTask_RejectsNonSeedableEnvBeforeDBConnection(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:1/annict?sslmode=disable&connect_timeout=1")
	t.Setenv("ANNICT_PORT", "3000")
	t.Setenv("ANNICT_DOMAIN", "example.com")
	t.Setenv("ANNICT_COOKIE_DOMAIN", "example.com")
	t.Setenv("ANNICT_SESSION_SECURE", "true")
	t.Setenv("ANNICT_SESSION_HTTPONLY", "true")
	t.Setenv("ANNICT_IMGPROXY_ENDPOINT", "https://example.com")
	t.Setenv("ANNICT_IMGPROXY_KEY", "00")
	t.Setenv("ANNICT_IMGPROXY_SALT", "00")

	err := tasks["seed"].run(context.Background())
	if err == nil {
		t.Fatal(`tasks["seed"].run() = nil、期待値 = エラーあり`)
	}
	if !strings.Contains(err.Error(), `APP_ENV="prod"`) {
		t.Fatalf(`tasks["seed"].run()のエラー = %q、期待値 = 環境ガードのエラー`, err)
	}
}

// TestSeed_RejectsNonSeedableEnvはseeder側のガードのテストと同様、意図的にnilの
// データベースハンドルを渡す。シード生成が走るかどうかを決めるのは本体が受け渡す環境であり、
// 拒否された場合はデータベースに到達する前に戻らなければならない。渡された設定を転送せず
// 自前で環境を読む実装であれば、ここではテスト環境が見えてガードを通過し、nilのハンドルで
// panicする。
func TestSeed_RejectsNonSeedableEnv(t *testing.T) {
	t.Parallel()

	if err := seed(context.Background(), &config.Config{Env: "prod"}, nil, nil); err == nil {
		t.Fatal("seed() = nil、期待値 = エラーあり")
	}
}
