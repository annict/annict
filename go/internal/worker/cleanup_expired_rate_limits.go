package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
)

// ExpiredRateLimitCleanerは終わった窓のレート制限のカウンターのクリーンアップを実行する。
type ExpiredRateLimitCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredRateLimitsWorkerは終わった窓のレート制限のカウンターのクリーンアップワーカー。
type CleanupExpiredRateLimitsWorker struct {
	river.WorkerDefaults[dispatcher.CleanupExpiredRateLimitsArgs]
	cleaner ExpiredRateLimitCleaner
}

// NewCleanupExpiredRateLimitsWorkerは新しいCleanupExpiredRateLimitsWorkerを作成する。
func NewCleanupExpiredRateLimitsWorker(cleaner ExpiredRateLimitCleaner) *CleanupExpiredRateLimitsWorker {
	return &CleanupExpiredRateLimitsWorker{
		cleaner: cleaner,
	}
}

// Workは終わった窓のレート制限のカウンターを削除する。
func (w *CleanupExpiredRateLimitsWorker) Work(ctx context.Context, job *river.Job[dispatcher.CleanupExpiredRateLimitsArgs]) error {
	return w.cleaner.Execute(ctx)
}
