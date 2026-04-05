package worker

import (
	"context"

	"github.com/riverqueue/river"
)

// CleanupExpiredTokensArgs はトークンクリーンアップジョブの引数です
type CleanupExpiredTokensArgs struct{}

// Kind はジョブの種類を返します
func (CleanupExpiredTokensArgs) Kind() string {
	return "cleanup_expired_tokens"
}

// InsertOpts はジョブ挿入時のデフォルトオプションを返します
func (CleanupExpiredTokensArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// ExpiredTokenCleaner は期限切れトークンのクリーンアップを実行するインターフェースです
type ExpiredTokenCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredTokensWorker はトークンクリーンアップワーカーです
type CleanupExpiredTokensWorker struct {
	river.WorkerDefaults[CleanupExpiredTokensArgs]
	cleaner ExpiredTokenCleaner
}

// NewCleanupExpiredTokensWorker は新しいCleanupExpiredTokensWorkerを作成します
func NewCleanupExpiredTokensWorker(cleaner ExpiredTokenCleaner) *CleanupExpiredTokensWorker {
	return &CleanupExpiredTokensWorker{
		cleaner: cleaner,
	}
}

// Work は有効期限切れおよび使用済みトークンを削除します
func (w *CleanupExpiredTokensWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredTokensArgs]) error {
	return w.cleaner.Execute(ctx)
}
