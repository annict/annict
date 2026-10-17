package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

// SetupTestRedisはテスト用Redisクライアントを返す。本関数ではDBをフラッシュしない。
// t.Parallel() と `go test ./...` の組み合わせにより複数パッケージが別プロセス
// として同じRedis DBを共有しているため、どこかでFlushDBすると実行中の
// 他テストのキーまで消えてしまう。各テストは、他テストと衝突しないキーを
// 使い、必要なら自分が触ったキーだけを後始末する責務を持つ。
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
