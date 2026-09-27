package worker

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
)

// SyncAnimesTimeoutはSyncAnimesWorkerの1回の試行の実行上限。Riverの既定の
// ジョブタイムアウト (1分) ではworks / episodesテーブル全体の走査が終わらず、
// context deadline exceededで打ち切られるため上書きする。ページごとに
// コミットした変更は次回の実行でも残るが、走査は先頭からやり直す。
// 1回の試行の上限は定期ジョブの投入間隔 (1時間) より短くする。
const SyncAnimesTimeout = 30 * time.Minute

// AnimesSyncerはworks/episodes -> animesのリコンサイルを実行する。
// 実体は *usecase.SyncAnimesUsecase。
type AnimesSyncer interface {
	Execute(ctx context.Context) (*usecase.SyncAnimesResult, error)
}

// SyncAnimesWorkerはフェーズ2のリコンサイルバッチの薄いRiverアダプタ。
type SyncAnimesWorker struct {
	river.WorkerDefaults[dispatcher.SyncAnimesArgs]
	syncer AnimesSyncer
}

// NewSyncAnimesWorkerはSyncAnimesWorkerを生成する。
func NewSyncAnimesWorker(syncer AnimesSyncer) *SyncAnimesWorker {
	return &SyncAnimesWorker{syncer: syncer}
}

// Timeoutはジョブの実行上限を返す。
func (w *SyncAnimesWorker) Timeout(*river.Job[dispatcher.SyncAnimesArgs]) time.Duration {
	return SyncAnimesTimeout
}

// Workはリコンサイルを実行する。件数はUseCase内でログ出力されるため、Workerは
// エラーをそのまま伝搬するだけにとどめ、ジョブ実行ログとリトライはRiverに任せる。
func (w *SyncAnimesWorker) Work(ctx context.Context, job *river.Job[dispatcher.SyncAnimesArgs]) error {
	_, err := w.syncer.Execute(ctx)
	return err
}
