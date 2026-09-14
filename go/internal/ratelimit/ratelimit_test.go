package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/testutil"
)

func TestLimiter_Check(t *testing.T) {
	t.Parallel()

	rdb := testutil.SetupTestRedis(t)
	limiter := NewLimiter(rdb)
	ctx := context.Background()

	// 同じRedis DBを共有する並列テストと干渉しないよう、本テストに
	// スコープしたキーを使う。
	prefix := testutil.UniqueRateLimitPrefix(t)

	tests := []struct {
		name        string
		key         string
		limit       int
		window      time.Duration
		attempts    int
		wantAllowed []bool
	}{
		{
			name:        "5回制限で5回まで許可される",
			key:         prefix + ":limit5",
			limit:       5,
			window:      1 * time.Hour,
			attempts:    6,
			wantAllowed: []bool{true, true, true, true, true, false},
		},
		{
			name:        "3回制限で3回まで許可される",
			key:         prefix + ":limit3",
			limit:       3,
			window:      1 * time.Hour,
			attempts:    5,
			wantAllowed: []bool{true, true, true, false, false},
		},
		{
			name:        "1回制限で1回のみ許可される",
			key:         prefix + ":limit1",
			limit:       1,
			window:      1 * time.Hour,
			attempts:    3,
			wantAllowed: []bool{true, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 共有Redisに前回の状態が残っていてもカウンタを0から
			// 始められるよう、本ケースのキーだけをResetする。
			if err := limiter.Reset(ctx, tt.key); err != nil {
				t.Fatalf("resetのエラー = %v", err)
			}
			t.Cleanup(func() { _ = limiter.Reset(ctx, tt.key) })

			for i := 0; i < tt.attempts; i++ {
				allowed, err := limiter.Check(ctx, tt.key, tt.limit, tt.window)
				if err != nil {
					t.Fatalf("%d回目の試行の想定外のエラー = %v", i+1, err)
				}

				if allowed != tt.wantAllowed[i] {
					t.Errorf("%d回目の試行のallowed = %v、期待値 = %v", i+1, allowed, tt.wantAllowed[i])
				}
			}
		})
	}
}

func TestLimiter_Reset(t *testing.T) {
	t.Parallel()

	rdb := testutil.SetupTestRedis(t)
	limiter := NewLimiter(rdb)
	ctx := context.Background()

	// 同じRedis DBを共有する並列テストと干渉しないよう、本テストに
	// スコープしたキーを使う。
	key := testutil.UniqueRateLimitPrefix(t) + ":reset"
	limit := 3
	window := 1 * time.Hour

	// 共有Redisに前回の状態が残っていてもカウンタを0から始められる
	// よう、このテストのキーだけをResetする。
	if err := limiter.Reset(ctx, key); err != nil {
		t.Fatalf("resetのエラー = %v", err)
	}
	t.Cleanup(func() { _ = limiter.Reset(ctx, key) })

	// 3回アクセスして制限に達する
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Check(ctx, key, limit, window)
		if err != nil {
			t.Fatalf("checkのエラー = %v", err)
		}
		if !allowed {
			t.Errorf("%d回目の試行が許可されなかった", i+1)
		}
	}

	// 4回目は制限される
	allowed, err := limiter.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("checkのエラー = %v", err)
	}
	if allowed {
		t.Error("4回目の試行がブロックされなかった")
	}

	// リセット
	if err := limiter.Reset(ctx, key); err != nil {
		t.Fatalf("resetのエラー = %v", err)
	}

	// リセット後は再び許可される
	allowed, err = limiter.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("リセット後の確認エラー = %v", err)
	}
	if !allowed {
		t.Error("リセット後の1回目の試行が許可されなかった")
	}
}

func TestLimiter_TTL(t *testing.T) {
	t.Parallel()

	rdb := testutil.SetupTestRedis(t)
	limiter := NewLimiter(rdb)
	ctx := context.Background()

	// 同じRedis DBを共有する並列テストと干渉しないよう、本テストに
	// スコープしたキーを使う。
	key := testutil.UniqueRateLimitPrefix(t) + ":ttl"
	limit := 5
	window := 2 * time.Second // 短い時間窓でテスト

	// 共有Redisに前回の状態が残っていてもカウンタを0から始められる
	// よう、このテストのキーだけをResetする。
	if err := limiter.Reset(ctx, key); err != nil {
		t.Fatalf("resetのエラー = %v", err)
	}
	t.Cleanup(func() { _ = limiter.Reset(ctx, key) })

	// 1回アクセス
	allowed, err := limiter.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("checkのエラー = %v", err)
	}
	if !allowed {
		t.Error("1回目の試行が許可されなかった")
	}

	// TTLが設定されているか確認
	ttl := rdb.TTL(ctx, "rate_limit:"+key).Val()
	if ttl <= 0 || ttl > window {
		t.Errorf("TTLの期待値 = 0より大きく%v以下、実測値 = %v", window, ttl)
	}

	// 時間窓を超えて待機
	time.Sleep(window + 100*time.Millisecond)

	// カウンタがリセットされているか確認
	allowed, err = limiter.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("TTL経過後の確認エラー = %v", err)
	}
	if !allowed {
		t.Error("TTL経過後の1回目の試行が許可されなかった")
	}
}
