package worker

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
)

// CleanupExpiredSessionsTimeoutはCleanupExpiredSessionsWorkerの1回の試行の実行
// 上限。Riverの既定のジョブタイムアウト (1分) ではUseCaseの上限時間より先に打ち切られ、
// context deadline exceededのエラーになるため上書きする。UseCaseは各バッチの後に
// 上限時間を判定するため、最後のバッチのぶん上限を超えうる。その余裕を足す。
// 前の実行と次の実行が重ならないよう、定期ジョブの投入間隔 (1時間) より短くする。
const CleanupExpiredSessionsTimeout = usecase.CleanupExpiredSessionsTimeLimit + 5*time.Minute

// ExpiredSessionCleanerは期限切れセッションのクリーンアップを実行する。
type ExpiredSessionCleaner interface {
	Execute(ctx context.Context) error
}

// CleanupExpiredSessionsWorkerはセッションクリーンアップワーカー。
type CleanupExpiredSessionsWorker struct {
	river.WorkerDefaults[dispatcher.CleanupExpiredSessionsArgs]
	cleaner ExpiredSessionCleaner
}

// NewCleanupExpiredSessionsWorkerは新しいCleanupExpiredSessionsWorkerを作成する。
func NewCleanupExpiredSessionsWorker(cleaner ExpiredSessionCleaner) *CleanupExpiredSessionsWorker {
	return &CleanupExpiredSessionsWorker{
		cleaner: cleaner,
	}
}

// Timeoutはジョブの実行上限を返す。
func (w *CleanupExpiredSessionsWorker) Timeout(*river.Job[dispatcher.CleanupExpiredSessionsArgs]) time.Duration {
	return CleanupExpiredSessionsTimeout
}

// Workは期限切れとなるまでアクセスされていないセッションを削除する。
func (w *CleanupExpiredSessionsWorker) Work(ctx context.Context, job *river.Job[dispatcher.CleanupExpiredSessionsArgs]) error {
	return w.cleaner.Execute(ctx)
}
