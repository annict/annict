package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
)

// ExpiredSessionCleanerは期限切れセッションのクリーンアップを実行する。
type ExpiredSessionCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredSessionsWorkerはセッションクリーンアップワーカー。
type CleanupExpiredSessionsWorker struct {
	river.WorkerDefaults[dispatcher.CleanupExpiredSessionsArgs]
	cleaner ExpiredSessionCleaner
}

// NewCleanupExpiredSessionsWorkerは新しいCleanupExpiredSessionsWorkerを作成する。
func NewCleanupExpiredSessionsWorker(cleaner ExpiredSessionCleaner) *CleanupExpiredSessionsWorker {
	return &CleanupExpiredSessionsWorker{
		cleaner: cleaner,
	}
}

// Workは期限切れとなるまでアクセスされていないセッションを削除する。
func (w *CleanupExpiredSessionsWorker) Work(ctx context.Context, job *river.Job[dispatcher.CleanupExpiredSessionsArgs]) error {
	return w.cleaner.Execute(ctx)
}
