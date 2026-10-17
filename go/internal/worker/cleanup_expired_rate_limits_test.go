package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/query"
	"github.com/annict/annict/go/internal/repository"
	"github.com/annict/annict/go/internal/testutil"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/worker"
)

func TestCleanupExpiredRateLimitsWorker(t *testing.T) {
	t.Parallel()

	db, tx := testutil.SetupTx(t)
	repo := repository.NewRateLimitRepository(query.New(db).WithTx(tx))
	w := worker.NewCleanupExpiredRateLimitsWorker(usecase.NewCleanupExpiredRateLimitsUsecase(repo))
	ctx := context.Background()

	key := testutil.UniqueRateLimitPrefix(t)
	if _, err := repo.Increment(ctx, key, time.Now().Add(-24*time.Hour)); err != nil {
		t.Fatalf("カウンターの作成に失敗: %v", err)
	}

	job := &river.Job[dispatcher.CleanupExpiredRateLimitsArgs]{
		Args: dispatcher.CleanupExpiredRateLimitsArgs{},
	}
	if err := w.Work(ctx, job); err != nil {
		t.Fatalf("ワーカーの実行に失敗: %v", err)
	}

	var count int64
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM rate_limits WHERE key = $1", key).Scan(&count); err != nil {
		t.Fatalf("カウンター数の取得に失敗: %v", err)
	}
	if count != 0 {
		t.Errorf("カウンター数 = %d、期待値 = 0", count)
	}
}
