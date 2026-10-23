package worker

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
)

// CleanupAnonymousSessionsTimeoutはCleanupAnonymousSessionsWorkerの1回の試行の実行
// 上限。CleanupExpiredSessionsTimeoutと同じく、Riverの既定のジョブタイムアウト (1分) を
// 上書きし、最後のバッチのぶんの余裕をUseCaseの上限時間に足す。前の実行と次の実行が
// 重ならないよう、定期ジョブの投入間隔 (1時間) より短くする。
const CleanupAnonymousSessionsTimeout = usecase.CleanupAnonymousSessionsTimeLimit + 5*time.Minute

// AnonymousSessionCleanerは未ログインのセッションのクリーンアップを実行する。
type AnonymousSessionCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupAnonymousSessionsWorkerは未ログインセッションのクリーンアップワーカー。
type CleanupAnonymousSessionsWorker struct {
	river.WorkerDefaults[dispatcher.CleanupAnonymousSessionsArgs]
	cleaner AnonymousSessionCleaner
}

// NewCleanupAnonymousSessionsWorkerは新しいCleanupAnonymousSessionsWorkerを作成する。
func NewCleanupAnonymousSessionsWorker(cleaner AnonymousSessionCleaner) *CleanupAnonymousSessionsWorker {
	return &CleanupAnonymousSessionsWorker{
		cleaner: cleaner,
	}
}

// Timeoutはジョブの実行上限を返す。
func (w *CleanupAnonymousSessionsWorker) Timeout(*river.Job[dispatcher.CleanupAnonymousSessionsArgs]) time.Duration {
	return CleanupAnonymousSessionsTimeout
}

// Workは最終アクセスから一定期間を過ぎた未ログインのセッションを削除する。
func (w *CleanupAnonymousSessionsWorker) Work(ctx context.Context, job *river.Job[dispatcher.CleanupAnonymousSessionsArgs]) error {
	return w.cleaner.Execute(ctx)
}
