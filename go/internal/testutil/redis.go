package testutil

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
)

// UniqueRateLimitPrefixは呼び出し元テストの名前から派生する一意なプレフィックスを返す。
// 同じRedis DBを共有する他テストとキーが衝突しないよう、Redisキー
// (またはハンドラーがRedisキーに組み込むemail / RemoteAddrなどの構成
// 値) を本プレフィックスで組み立てる。戻り値はRedisキーとメールアドレス
// のローカルパートのどちらでも安全に使える文字のみを含む: [A-Za-z0-9_]
// 以外の文字 (サブテストの区切り、空白、マルチバイト文字など) はすべて
// '_' に置換するため、日本語や記号を含むサブテスト名でも安全なprefixを
// 得られる。
func UniqueRateLimitPrefix(t *testing.T) string {
	t.Helper()
	// [A-Za-z0-9_] 以外のruneをすべて '_' に置換し、Redisキーの
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

// SetupTestRedisはテスト用Redisクライアントを返す。本関数ではDBをフラッシュしない。
// t.Parallel() と `go test ./...` の組み合わせにより複数パッケージが別プロセス
// として同じRedis DBを共有しているため、どこかでFlushDBすると実行中の
// 他テストのキーまで消えてしまう。各テストは、他テストと衝突しないキーを
// 使い、必要なら自分が触ったキーだけを後始末する責務を持つ
// (例: `limiter.Reset(ctx, key)`)。
func SetupTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	// 環境変数ANNICT_REDIS_URLから接続情報を取得する。
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
