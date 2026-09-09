package worker

import (
	"context"

	"github.com/riverqueue/river"
)

// CleanupExpiredSignInCodesArgs holds the arguments of the sign-in code cleanup job.
//
// [Ja] CleanupExpiredSignInCodesArgs は期限切れログインコードのクリーンアップジョブの引数。
type CleanupExpiredSignInCodesArgs struct{}

// Kind returns the job kind.
//
// [Ja] Kind はジョブの種類を返す。
func (CleanupExpiredSignInCodesArgs) Kind() string {
	return "cleanup_expired_sign_in_codes"
}

// InsertOpts returns the default options applied when the job is inserted.
//
// [Ja] InsertOpts はジョブ挿入時のデフォルトオプションを返す。
func (CleanupExpiredSignInCodesArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// ExpiredSignInCodeCleaner runs the cleanup of expired sign-in codes.
//
// [Ja] ExpiredSignInCodeCleaner は期限切れログインコードのクリーンアップを実行する。
type ExpiredSignInCodeCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredSignInCodesWorker is the worker that runs the sign-in code cleanup.
//
// [Ja] CleanupExpiredSignInCodesWorker は期限切れログインコードのクリーンアップワーカー。
type CleanupExpiredSignInCodesWorker struct {
	river.WorkerDefaults[CleanupExpiredSignInCodesArgs]
	cleaner ExpiredSignInCodeCleaner
}

// NewCleanupExpiredSignInCodesWorker creates a new CleanupExpiredSignInCodesWorker.
//
// [Ja] NewCleanupExpiredSignInCodesWorker は新しい CleanupExpiredSignInCodesWorker を作成する。
func NewCleanupExpiredSignInCodesWorker(cleaner ExpiredSignInCodeCleaner) *CleanupExpiredSignInCodesWorker {
	return &CleanupExpiredSignInCodesWorker{
		cleaner: cleaner,
	}
}

// Work deletes sign-in codes that have expired or have already been used.
//
// [Ja] Work は有効期限切れおよび使用済みのログインコードを削除する。
func (w *CleanupExpiredSignInCodesWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredSignInCodesArgs]) error {
	return w.cleaner.Execute(ctx)
}
