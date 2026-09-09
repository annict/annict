package main

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/config"
)

// TestSeedTask_Registered checks that the seed task name resolves to its environment
// guard and body. The name is the only thing an operator types, so the registration has
// to preserve both the pre-database safety check and the requested operation.
//
// [Ja] TestSeedTask_Registered は seed というタスク名が環境ガードと本体に解決することを確認
// する。運用者が打ち込むのはタスク名だけであるため、登録は DB 接続前の安全確認と要求された
// 処理の両方を保たなければならない。
func TestSeedTask_Registered(t *testing.T) {
	t.Parallel()

	task, ok := tasks["seed"]
	if !ok {
		t.Fatal(`task "seed" is not registered`)
	}
	if task.body == nil {
		t.Fatal(`task "seed" has no body`)
	}
	if task.guard == nil {
		t.Fatal(`task "seed" has no guard`)
	}
	if got, want := reflect.ValueOf(task.guard).Pointer(), reflect.ValueOf(guardSeed).Pointer(); got != want {
		t.Error(`task "seed" is registered with the wrong guard`)
	}
	if got, want := reflect.ValueOf(task.body).Pointer(), reflect.ValueOf(seed).Pointer(); got != want {
		t.Error(`task "seed" is registered with the wrong body`)
	}
}

// TestSeedTask_RejectsNonSeedableEnvBeforeDBConnection runs the registered task with
// an unreachable database on purpose. The production environment must be rejected by
// the task guard first; reaching sql.Open or PingContext would replace the environment
// error with a connection error and fail this test.
//
// This test cannot run in parallel because config.Load reads process environment
// variables, which t.Setenv changes for the duration of the test.
//
// [Ja] TestSeedTask_RejectsNonSeedableEnvBeforeDBConnection は、意図的に接続できない
// データベースを設定して登録済みタスクを実行する。本番環境はタスクガードが先に拒否しなければ
// ならない。sql.Open や PingContext まで到達すると環境エラーが接続エラーに置き換わり、本テスト
// は失敗する。
//
// config.Load はプロセスの環境変数を読み、t.Setenv がテスト中にその値を変えるため、本テストは
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
		t.Fatal(`tasks["seed"].run() = nil, want error`)
	}
	if !strings.Contains(err.Error(), `APP_ENV="prod"`) {
		t.Fatalf(`tasks["seed"].run() error = %q, want the environment guard error`, err)
	}
}

// TestSeed_RejectsNonSeedableEnv passes a nil database handle on purpose, the way the
// seeder's own guard test does: the environment the body forwards is what decides
// whether seeding runs, and a rejected one must return before anything reaches the
// database. A body that read the environment on its own instead of forwarding the
// configuration it was handed would see the test environment here, pass the guard and
// panic on the nil handle.
//
// [Ja] TestSeed_RejectsNonSeedableEnv は seeder 側のガードのテストと同様、意図的に nil の
// データベースハンドルを渡す。シード生成が走るかどうかを決めるのは本体が受け渡す環境であり、
// 拒否された場合はデータベースに到達する前に戻らなければならない。渡された設定を転送せず
// 自前で環境を読む実装であれば、ここではテスト環境が見えてガードを通過し、nil のハンドルで
// panic する。
func TestSeed_RejectsNonSeedableEnv(t *testing.T) {
	t.Parallel()

	if err := seed(context.Background(), &config.Config{Env: "prod"}, nil, nil); err == nil {
		t.Fatal("seed() = nil, want error")
	}
}
