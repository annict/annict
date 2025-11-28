package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/internal/testutil"
)

func TestLimiter_Check(t *testing.T) {
	rdb := testutil.SetupTestRedis(t)
	limiter := NewLimiter(rdb)
	ctx := context.Background()

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
			key:         "test:limit5",
			limit:       5,
			window:      1 * time.Hour,
			attempts:    6,
			wantAllowed: []bool{true, true, true, true, true, false},
		},
		{
			name:        "3回制限で3回まで許可される",
			key:         "test:limit3",
			limit:       3,
			window:      1 * time.Hour,
			attempts:    5,
			wantAllowed: []bool{true, true, true, false, false},
		},
		{
			name:        "1回制限で1回のみ許可される",
			key:         "test:limit1",
			limit:       1,
			window:      1 * time.Hour,
			attempts:    3,
			wantAllowed: []bool{true, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.attempts; i++ {
				allowed, err := limiter.Check(ctx, tt.key, tt.limit, tt.window)
				if err != nil {
					t.Fatalf("attempt %d: unexpected error: %v", i+1, err)
				}

				if allowed != tt.wantAllowed[i] {
					t.Errorf("attempt %d: allowed = %v, want %v", i+1, allowed, tt.wantAllowed[i])
				}
			}
		})
	}
}

func TestLimiter_Reset(t *testing.T) {
	rdb := testutil.SetupTestRedis(t)
	limiter := NewLimiter(rdb)
	ctx := context.Background()

	key := "test:reset"
	limit := 3
	window := 1 * time.Hour

	// 3回アクセスして制限に達する
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Check(ctx, key, limit, window)
		if err != nil {
			t.Fatalf("check failed: %v", err)
		}
		if !allowed {
			t.Errorf("attempt %d should be allowed", i+1)
		}
	}

	// 4回目は制限される
	allowed, err := limiter.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	if allowed {
		t.Error("4th attempt should be blocked")
	}

	// リセット
	if err := limiter.Reset(ctx, key); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	// リセット後は再び許可される
	allowed, err = limiter.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("check after reset failed: %v", err)
	}
	if !allowed {
		t.Error("first attempt after reset should be allowed")
	}
}

func TestLimiter_TTL(t *testing.T) {
	rdb := testutil.SetupTestRedis(t)
	limiter := NewLimiter(rdb)
	ctx := context.Background()

	key := "test:ttl"
	limit := 5
	window := 2 * time.Second // 短い時間窓でテスト

	// 1回アクセス
	allowed, err := limiter.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	if !allowed {
		t.Error("first attempt should be allowed")
	}

	// TTLが設定されているか確認
	ttl := rdb.TTL(ctx, "rate_limit:"+key).Val()
	if ttl <= 0 || ttl > window {
		t.Errorf("TTL should be between 0 and %v, got %v", window, ttl)
	}

	// 時間窓を超えて待機
	time.Sleep(window + 100*time.Millisecond)

	// カウンタがリセットされているか確認
	allowed, err = limiter.Check(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("check after TTL failed: %v", err)
	}
	if !allowed {
		t.Error("first attempt after TTL should be allowed")
	}
}
