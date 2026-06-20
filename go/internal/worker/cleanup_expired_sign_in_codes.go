package worker

import (
	"context"

	"github.com/riverqueue/river"
)

// CleanupExpiredSignInCodesArgs は期限切れログインコードのクリーンアップジョブの引数です
type CleanupExpiredSignInCodesArgs struct{}

// Kind はジョブの種類を返します
func (CleanupExpiredSignInCodesArgs) Kind() string {
	return "cleanup_expired_sign_in_codes"
}

// InsertOpts はジョブ挿入時のデフォルトオプションを返します
func (CleanupExpiredSignInCodesArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// ExpiredSignInCodeCleaner は期限切れログインコードのクリーンアップを実行するインターフェースです
type ExpiredSignInCodeCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredSignInCodesWorker は期限切れログインコードのクリーンアップワーカーです
type CleanupExpiredSignInCodesWorker struct {
	river.WorkerDefaults[CleanupExpiredSignInCodesArgs]
	cleaner ExpiredSignInCodeCleaner
}

// NewCleanupExpiredSignInCodesWorker は新しいCleanupExpiredSignInCodesWorkerを作成します
func NewCleanupExpiredSignInCodesWorker(cleaner ExpiredSignInCodeCleaner) *CleanupExpiredSignInCodesWorker {
	return &CleanupExpiredSignInCodesWorker{
		cleaner: cleaner,
	}
}

// Work は有効期限切れおよび使用済みのログインコードを削除します
func (w *CleanupExpiredSignInCodesWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredSignInCodesArgs]) error {
	return w.cleaner.Execute(ctx)
}
