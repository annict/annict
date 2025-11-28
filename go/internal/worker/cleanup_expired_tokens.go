package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/annict/annict/internal/query"
	"github.com/riverqueue/river"
)

// CleanupExpiredTokensArgs はトークンクリーンアップジョブの引数です
type CleanupExpiredTokensArgs struct{}

// Kind はジョブの種類を返します
func (CleanupExpiredTokensArgs) Kind() string {
	return "cleanup_expired_tokens"
}

// CleanupExpiredTokensWorker はトークンクリーンアップワーカーです
type CleanupExpiredTokensWorker struct {
	river.WorkerDefaults[CleanupExpiredTokensArgs]
	queries *query.Queries
}

// NewCleanupExpiredTokensWorker は新しいCleanupExpiredTokensWorkerを作成します
func NewCleanupExpiredTokensWorker(queries *query.Queries) *CleanupExpiredTokensWorker {
	return &CleanupExpiredTokensWorker{
		queries: queries,
	}
}

// Work は有効期限切れおよび使用済みトークンを削除します
func (w *CleanupExpiredTokensWorker) Work(ctx context.Context, job *river.Job[CleanupExpiredTokensArgs]) error {
	slog.InfoContext(ctx, "トークンクリーンアップジョブを開始します")

	// 24時間以上前に期限切れまたは使用済みになったトークンを削除
	cutoff := time.Now().Add(-24 * time.Hour)

	err := w.queries.DeleteExpiredPasswordResetTokens(ctx, cutoff)
	if err != nil {
		slog.ErrorContext(ctx, "トークンの削除に失敗しました",
			"cutoff", cutoff,
			"error", err,
		)
		return err
	}

	slog.InfoContext(ctx, "トークンクリーンアップジョブが完了しました",
		"cutoff", cutoff,
	)

	return nil
}
