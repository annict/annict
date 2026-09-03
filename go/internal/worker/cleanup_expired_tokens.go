package worker

import (
	"context"

	"github.com/riverqueue/river"
)

// CleanupExpiredTokensArgs holds the arguments of the token cleanup job.
//
// [Ja] CleanupExpiredTokensArgs はトークンクリーンアップジョブの引数。
type CleanupExpiredTokensArgs struct{}

// Kind returns the job kind.
//
// [Ja] Kind はジョブの種類を返す。
func (CleanupExpiredTokensArgs) Kind() string {
	return "cleanup_expired_tokens"
}

// InsertOpts returns the default options applied when the job is inserted.
//
// [Ja] InsertOpts はジョブ挿入時のデフォルトオプションを返す。
func (CleanupExpiredTokensArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// ExpiredTokenCleaner runs the cleanup of expired tokens.
//
// [Ja] ExpiredTokenCleaner は期限切れトークンのクリーンアップを実行する。
type ExpiredTokenCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredTokensWorker is the worker that runs the token cleanup.
//
// [Ja] CleanupExpiredTokensWorker はトークンクリーンアップワーカー。
type CleanupExpiredTokensWorker struct {
	river.WorkerDefaults[CleanupExpiredTokensArgs]
	cleaner ExpiredTokenCleaner
}

// NewCleanupExpiredTokensWorker creates a new CleanupExpiredTokensWorker.
//
// [Ja] NewCleanupExpiredTokensWorker は新しい CleanupExpiredTokensWorker を作成する。
func NewCleanupExpiredTokensWorker(cleaner ExpiredTokenCleaner) *CleanupExpiredTokensWorker {
	return &CleanupExpiredTokensWorker{
		cleaner: cleaner,
	}
}

// Work deletes tokens that have expired or have already been used.
//
// [Ja] Work は有効期限切れおよび使用済みのトークンを削除する。
func (w *CleanupExpiredTokensWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredTokensArgs]) error {
	return w.cleaner.Execute(ctx)
}
