package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

// newTestLimiterはトランザクション内で数え、現在時刻をnowに固定した限定器を返す
func newTestLimiter(t *testing.T, now *time.Time) *Limiter {
	t.Helper()

	db, tx := testutil.SetupTx(t)
	limiter := NewLimiter(repository.NewRateLimitRepository(query.New(db))).WithTx(tx)
	limiter.now = func() time.Time { return *now }
	return limiter
}

func TestLimiter_Check(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		limit       int
		attempts    int
		wantAllowed []bool
	}{
		{
			name:        "5回制限で5回まで許可される",
			limit:       5,
			attempts:    6,
			wantAllowed: []bool{true, true, true, true, true, false},
		},
		{
			name:        "1回制限で1回のみ許可される",
			limit:       1,
			attempts:    3,
			wantAllowed: []bool{true, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			now := time.Date(2026, 10, 14, 12, 34, 56, 0, time.UTC)
			limiter := newTestLimiter(t, &now)
			input := CheckInput{Key: testutil.UniqueRateLimitPrefix(t), Limit: tt.limit, Window: time.Hour}

			for i := 0; i < tt.attempts; i++ {
				result, err := limiter.Check(context.Background(), input)
				if err != nil {
					t.Fatalf("%d回目の試行の想定外のエラー = %v", i+1, err)
				}
				if result.Allowed != tt.wantAllowed[i] {
					t.Errorf("%d回目の試行のAllowed = %v、期待値 = %v", i+1, result.Allowed, tt.wantAllowed[i])
				}
				if result.Count != i+1 {
					t.Errorf("%d回目の試行のCount = %d、期待値 = %d", i+1, result.Count, i+1)
				}
				if wantRemaining := max(tt.limit-(i+1), 0); result.Remaining != wantRemaining {
					t.Errorf("%d回目の試行のRemaining = %d、期待値 = %d", i+1, result.Remaining, wantRemaining)
				}
			}
		})
	}
}

// TestLimiter_Check_Windowは窓が時刻の切り捨てで決まり、次の窓では0から数え直すことを確認する
func TestLimiter_Check_Window(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 14, 12, 34, 56, 0, time.UTC)
	limiter := newTestLimiter(t, &now)
	input := CheckInput{Key: testutil.UniqueRateLimitPrefix(t), Limit: 1, Window: time.Minute}
	ctx := context.Background()

	first, err := limiter.Check(ctx, input)
	if err != nil {
		t.Fatalf("1回目のチェックのエラー = %v", err)
	}
	if want := time.Date(2026, 10, 14, 12, 35, 0, 0, time.UTC); !first.ResetAt.Equal(want) {
		t.Errorf("ResetAt = %v、期待値 = %v", first.ResetAt, want)
	}

	// 同じ窓の終わりぎわ (12:34:59) は上限を超える
	now = time.Date(2026, 10, 14, 12, 34, 59, 0, time.UTC)
	second, err := limiter.Check(ctx, input)
	if err != nil {
		t.Fatalf("2回目のチェックのエラー = %v", err)
	}
	if second.Allowed {
		t.Error("同じ窓の2回目が許可された")
	}

	// 次の窓 (12:35:00) では数え直す
	now = time.Date(2026, 10, 14, 12, 35, 0, 0, time.UTC)
	third, err := limiter.Check(ctx, input)
	if err != nil {
		t.Fatalf("3回目のチェックのエラー = %v", err)
	}
	if !third.Allowed || third.Count != 1 {
		t.Errorf("次の窓のAllowed = %v、Count = %d、期待値 = true、1", third.Allowed, third.Count)
	}
}

func TestLimiter_Check_InvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input CheckInput
	}{
		{name: "キーが空", input: CheckInput{Key: "", Limit: 1, Window: time.Minute}},
		{name: "上限が0", input: CheckInput{Key: "key", Limit: 0, Window: time.Minute}},
		{name: "窓が0", input: CheckInput{Key: "key", Limit: 1, Window: 0}},
		{name: "窓がMaxWindowより長い", input: CheckInput{Key: "key", Limit: 1, Window: MaxWindow + time.Second}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
			limiter := newTestLimiter(t, &now)
			if _, err := limiter.Check(context.Background(), tt.input); err == nil {
				t.Error("エラーが返らなかった")
			}
		})
	}
}
