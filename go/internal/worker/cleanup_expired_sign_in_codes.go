package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
)

// ExpiredSignInCodeCleanerは期限切れログインコードのクリーンアップを実行する。
type ExpiredSignInCodeCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredSignInCodesWorkerは期限切れログインコードのクリーンアップワーカー。
type CleanupExpiredSignInCodesWorker struct {
	river.WorkerDefaults[dispatcher.CleanupExpiredSignInCodesArgs]
	cleaner ExpiredSignInCodeCleaner
}

// NewCleanupExpiredSignInCodesWorkerは新しいCleanupExpiredSignInCodesWorkerを作成する。
func NewCleanupExpiredSignInCodesWorker(cleaner ExpiredSignInCodeCleaner) *CleanupExpiredSignInCodesWorker {
	return &CleanupExpiredSignInCodesWorker{
		cleaner: cleaner,
	}
}

// Workは有効期限切れおよび使用済みのログインコードを削除する。
func (w *CleanupExpiredSignInCodesWorker) Work(ctx context.Context, job *river.Job[dispatcher.CleanupExpiredSignInCodesArgs]) error {
	return w.cleaner.Execute(ctx)
}
