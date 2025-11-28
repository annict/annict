package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

// SetupTestRedis はテスト用 Redis クライアントをセットアップし、
// テスト終了時に自動的にクリーンアップします
func SetupTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	// 環境変数 ANNICT_REDIS_URL から接続情報を取得
	// CI環境: redis://localhost:6379/1
	// Dev Container: redis://redis:6379/1
	redisURL := os.Getenv("ANNICT_REDIS_URL")
	if redisURL == "" {
		// 環境変数が設定されていない場合は、デフォルト値を使用（Dev Container環境）
		// docker-compose.ymlで定義されているredisサービスに接続
		redisURL = "redis://redis:6379/1"
	}

	// Redis URLをパース
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("failed to parse Redis URL: %v", err)
	}

	// テスト用 Redis に接続
	rdb := redis.NewClient(opts)

	// 接続確認
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to connect to test Redis: %v", err)
	}

	// テスト終了時にデータをクリーンアップ
	t.Cleanup(func() {
		// テストで使用したキーをすべて削除
		_ = rdb.FlushDB(ctx).Err()
		_ = rdb.Close()
	})

	return rdb
}
