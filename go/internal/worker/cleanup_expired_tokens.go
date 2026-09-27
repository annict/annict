package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
)

// ExpiredTokenCleanerは期限切れトークンのクリーンアップを実行する。
type ExpiredTokenCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredTokensWorkerはトークンクリーンアップワーカー。
type CleanupExpiredTokensWorker struct {
	river.WorkerDefaults[dispatcher.CleanupExpiredTokensArgs]
	cleaner ExpiredTokenCleaner
}

// NewCleanupExpiredTokensWorkerは新しいCleanupExpiredTokensWorkerを作成する。
func NewCleanupExpiredTokensWorker(cleaner ExpiredTokenCleaner) *CleanupExpiredTokensWorker {
	return &CleanupExpiredTokensWorker{
		cleaner: cleaner,
	}
}

// Workは有効期限切れおよび使用済みのトークンを削除する。
func (w *CleanupExpiredTokensWorker) Work(ctx context.Context, job *river.Job[dispatcher.CleanupExpiredTokensArgs]) error {
	return w.cleaner.Execute(ctx)
}
