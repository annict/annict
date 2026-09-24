package worker

import (
	"context"

	"github.com/riverqueue/river"
)

// CleanupExpiredSessionsArgsは期限切れセッションのクリーンアップジョブの引数。
type CleanupExpiredSessionsArgs struct{}

// Kindはジョブの種類を返す。
func (CleanupExpiredSessionsArgs) Kind() string {
	return "cleanup_expired_sessions"
}

// InsertOptsはジョブ挿入時のデフォルトオプションを返す。
func (CleanupExpiredSessionsArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// ExpiredSessionCleanerは期限切れセッションのクリーンアップを実行する。
type ExpiredSessionCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredSessionsWorkerはセッションクリーンアップワーカー。
type CleanupExpiredSessionsWorker struct {
	river.WorkerDefaults[CleanupExpiredSessionsArgs]
	cleaner ExpiredSessionCleaner
}

// NewCleanupExpiredSessionsWorkerは新しいCleanupExpiredSessionsWorkerを作成する。
func NewCleanupExpiredSessionsWorker(cleaner ExpiredSessionCleaner) *CleanupExpiredSessionsWorker {
	return &CleanupExpiredSessionsWorker{
		cleaner: cleaner,
	}
}

// Workは期限切れとなるまでアクセスされていないセッションを削除する。
func (w *CleanupExpiredSessionsWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredSessionsArgs]) error {
	return w.cleaner.Execute(ctx)
}
