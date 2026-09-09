package worker

import (
	"context"

	"github.com/riverqueue/river"
)

// CleanupExpiredSessionsArgs holds the arguments of the session cleanup job.
//
// [Ja] CleanupExpiredSessionsArgs は期限切れセッションのクリーンアップジョブの引数。
type CleanupExpiredSessionsArgs struct{}

// Kind returns the job kind.
//
// [Ja] Kind はジョブの種類を返す。
func (CleanupExpiredSessionsArgs) Kind() string {
	return "cleanup_expired_sessions"
}

// InsertOpts returns the default options applied when the job is inserted.
//
// [Ja] InsertOpts はジョブ挿入時のデフォルトオプションを返す。
func (CleanupExpiredSessionsArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// ExpiredSessionCleaner runs the cleanup of expired sessions.
//
// [Ja] ExpiredSessionCleaner は期限切れセッションのクリーンアップを実行する。
type ExpiredSessionCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredSessionsWorker is the worker that runs the session cleanup.
//
// [Ja] CleanupExpiredSessionsWorker はセッションクリーンアップワーカー。
type CleanupExpiredSessionsWorker struct {
	river.WorkerDefaults[CleanupExpiredSessionsArgs]
	cleaner ExpiredSessionCleaner
}

// NewCleanupExpiredSessionsWorker creates a new CleanupExpiredSessionsWorker.
//
// [Ja] NewCleanupExpiredSessionsWorker は新しい CleanupExpiredSessionsWorker を作成する。
func NewCleanupExpiredSessionsWorker(cleaner ExpiredSessionCleaner) *CleanupExpiredSessionsWorker {
	return &CleanupExpiredSessionsWorker{
		cleaner: cleaner,
	}
}

// Work deletes the sessions that have not been accessed for long enough to expire.
//
// [Ja] Work は期限切れとなるまでアクセスされていないセッションを削除する。
func (w *CleanupExpiredSessionsWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredSessionsArgs]) error {
	return w.cleaner.Execute(ctx)
}
