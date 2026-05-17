package testutil

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
)

// UniqueRateLimitPrefix returns a prefix derived from the calling test's
// name. Compose Redis keys (or values that feed into Redis keys, like the
// email address or RemoteAddr host) with this prefix so that the test's
// keys never collide with other tests sharing the same Redis DB. The
// returned string contains only characters that are valid in both Redis
// keys and the local part of an email address: any rune outside
// [A-Za-z0-9_] (subtest separators, whitespace, multibyte characters, etc.)
// is translated to '_' so that subtest names written in Japanese or with
// punctuation still produce a safe prefix.
//
// [Ja] 呼び出し元テストの名前から派生する一意なプレフィックスを返す。
// 同じ Redis DB を共有する他テストとキーが衝突しないよう、Redis キー
// (またはハンドラーが Redis キーに組み込む email / RemoteAddr などの構成
// 値) を本プレフィックスで組み立てる。戻り値は Redis キーとメールアドレス
// のローカルパートのどちらでも安全に使える文字のみを含む: [A-Za-z0-9_]
// 以外の文字 (サブテストの区切り、空白、マルチバイト文字など) はすべて
// '_' に置換するため、日本語や記号を含むサブテスト名でも安全な prefix を
// 得られる。
func UniqueRateLimitPrefix(t *testing.T) string {
	t.Helper()
	// Translate every rune outside [A-Za-z0-9_] to '_' so the result is
	// safe both as a Redis key fragment and as the local part of an email
	// address. Subtests separate their parent and child name with '/'
	// (e.g. "TestFoo/case_a"); this rule covers that case as well.
	//
	// [Ja] [A-Za-z0-9_] 以外の rune をすべて '_' に置換し、Redis キーの
	// 一部としてもメールアドレスのローカルパートとしても安全な文字列に
	// する。サブテストの '/' (例: "TestFoo/case_a") もこのルールに含まれる。
	name := t.Name()
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// SetupTestRedis returns a Redis client connected to the test Redis instance.
// It never flushes the DB: with t.Parallel() and `go test ./...` running
// multiple packages as separate processes, all sharing the same Redis DB,
// any FlushDB call would wipe keys that other in-flight tests still depend
// on. Each test is responsible for using keys unique enough that it does
// not collide with other tests, and for cleaning up only its own keys (e.g.
// via `limiter.Reset(ctx, key)`).
//
// [Ja] テスト用 Redis クライアントを返す。本関数では DB をフラッシュしない。
// t.Parallel() と `go test ./...` の組み合わせにより複数パッケージが別プロセス
// として同じ Redis DB を共有しているため、どこかで FlushDB すると実行中の
// 他テストのキーまで消えてしまう。各テストは、他テストと衝突しないキーを
// 使い、必要なら自分が触ったキーだけを後始末する責務を持つ
// (例: `limiter.Reset(ctx, key)`)。
func SetupTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	// Read connection info from ANNICT_REDIS_URL.
	// CI: redis://localhost:6379/1, Dev Container: redis://redis:6379/1.
	//
	// [Ja] 環境変数 ANNICT_REDIS_URL から接続情報を取得する。
	// CI: redis://localhost:6379/1, Dev Container: redis://redis:6379/1。
	redisURL := os.Getenv("ANNICT_REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379/1"
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("failed to parse Redis URL: %v", err)
	}

	rdb := redis.NewClient(opts)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("failed to connect to test Redis: %v", err)
	}

	t.Cleanup(func() {
		_ = rdb.Close()
	})

	return rdb
}
