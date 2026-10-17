package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/ratelimit"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
)

func TestCleanupExpiredRateLimitsUsecase_Execute(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewRateLimitRepository(query.New(db).WithTx(tx))
	uc := NewCleanupExpiredRateLimitsUsecase(repo)
	ctx := context.Background()

	key := testutil.UniqueRateLimitPrefix(t)
	now := time.Now().Truncate(time.Second)
	// 1分ずらして、テストの実行中に時刻が進んでもcutoffをまたがないようにする
	expired := now.Add(-ratelimit.MaxWindow - time.Minute)
	active := now.Add(-ratelimit.MaxWindow + time.Minute)
	for _, windowStart := range []time.Time{expired, active} {
		if _, err := repo.Increment(ctx, key, windowStart); err != nil {
			t.Fatalf("カウンターの作成に失敗: %v", err)
		}
	}

	if err := uc.Execute(ctx); err != nil {
		t.Fatalf("Executeのエラー = %v", err)
	}

	var remaining []time.Time
	rows, err := tx.QueryContext(ctx, "SELECT window_start FROM rate_limits WHERE key = $1", key)
	if err != nil {
		t.Fatalf("カウンターの取得に失敗: %v", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var windowStart time.Time
		if err := rows.Scan(&windowStart); err != nil {
			t.Fatalf("カウンターの読み取りに失敗: %v", err)
		}
		remaining = append(remaining, windowStart)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("カウンターの読み取りに失敗: %v", err)
	}

	if len(remaining) != 1 || !remaining[0].Equal(active) {
		t.Errorf("残ったカウンターの窓 = %v、期待値 = [%v]", remaining, active)
	}
}
