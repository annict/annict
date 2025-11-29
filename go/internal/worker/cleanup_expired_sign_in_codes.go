package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/annict/annict/internal/query"
	"github.com/riverqueue/river"
)

// CleanupExpiredSignInCodesArgs は期限切れログインコードのクリーンアップジョブの引数です
type CleanupExpiredSignInCodesArgs struct{}

// Kind はジョブの種類を返します
func (CleanupExpiredSignInCodesArgs) Kind() string {
	return "cleanup_expired_sign_in_codes"
}

// CleanupExpiredSignInCodesWorker は期限切れログインコードのクリーンアップワーカーです
type CleanupExpiredSignInCodesWorker struct {
	river.WorkerDefaults[CleanupExpiredSignInCodesArgs]
	queries *query.Queries
}

// NewCleanupExpiredSignInCodesWorker は新しいCleanupExpiredSignInCodesWorkerを作成します
func NewCleanupExpiredSignInCodesWorker(queries *query.Queries) *CleanupExpiredSignInCodesWorker {
	return &CleanupExpiredSignInCodesWorker{
		queries: queries,
	}
}

// Work は有効期限切れおよび使用済みのログインコードを削除します
func (w *CleanupExpiredSignInCodesWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredSignInCodesArgs]) error {
	slog.InfoContext(ctx, "ログインコードクリーンアップジョブを開始します")

	// 24時間以上前に期限切れまたは使用済みになったコードを削除
	cutoff := time.Now().Add(-24 * time.Hour)

	err := w.queries.DeleteExpiredSignInCodes(ctx, cutoff)
	if err != nil {
		slog.ErrorContext(ctx, "ログインコードの削除に失敗しました",
			"cutoff", cutoff,
			"error", err,
		)
		return err
	}

	slog.InfoContext(ctx, "ログインコードクリーンアップジョブが完了しました",
		"cutoff", cutoff,
	)

	return nil
}
