package worker

import (
	"context"

	"github.com/riverqueue/river"
)

// CleanupExpiredTokensArgsはトークンクリーンアップジョブの引数。
type CleanupExpiredTokensArgs struct{}

// Kindはジョブの種類を返す。
func (CleanupExpiredTokensArgs) Kind() string {
	return "cleanup_expired_tokens"
}

// InsertOptsはジョブ挿入時のデフォルトオプションを返す。
func (CleanupExpiredTokensArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// ExpiredTokenCleanerは期限切れトークンのクリーンアップを実行する。
type ExpiredTokenCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredTokensWorkerはトークンクリーンアップワーカー。
type CleanupExpiredTokensWorker struct {
	river.WorkerDefaults[CleanupExpiredTokensArgs]
	cleaner ExpiredTokenCleaner
}

// NewCleanupExpiredTokensWorkerは新しいCleanupExpiredTokensWorkerを作成する。
func NewCleanupExpiredTokensWorker(cleaner ExpiredTokenCleaner) *CleanupExpiredTokensWorker {
	return &CleanupExpiredTokensWorker{
		cleaner: cleaner,
	}
}

// Workは有効期限切れおよび使用済みのトークンを削除する。
func (w *CleanupExpiredTokensWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredTokensArgs]) error {
	return w.cleaner.Execute(ctx)
}
