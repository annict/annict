package worker

import (
	"context"

	"github.com/riverqueue/river"
)

// CleanupExpiredSignInCodesArgsは期限切れログインコードのクリーンアップジョブの引数。
type CleanupExpiredSignInCodesArgs struct{}

// Kindはジョブの種類を返す。
func (CleanupExpiredSignInCodesArgs) Kind() string {
	return "cleanup_expired_sign_in_codes"
}

// InsertOptsはジョブ挿入時のデフォルトオプションを返す。
func (CleanupExpiredSignInCodesArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// ExpiredSignInCodeCleanerは期限切れログインコードのクリーンアップを実行する。
type ExpiredSignInCodeCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredSignInCodesWorkerは期限切れログインコードのクリーンアップワーカー。
type CleanupExpiredSignInCodesWorker struct {
	river.WorkerDefaults[CleanupExpiredSignInCodesArgs]
	cleaner ExpiredSignInCodeCleaner
}

// NewCleanupExpiredSignInCodesWorkerは新しいCleanupExpiredSignInCodesWorkerを作成する。
func NewCleanupExpiredSignInCodesWorker(cleaner ExpiredSignInCodeCleaner) *CleanupExpiredSignInCodesWorker {
	return &CleanupExpiredSignInCodesWorker{
		cleaner: cleaner,
	}
}

// Workは有効期限切れおよび使用済みのログインコードを削除する。
func (w *CleanupExpiredSignInCodesWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredSignInCodesArgs]) error {
	return w.cleaner.Execute(ctx)
}
