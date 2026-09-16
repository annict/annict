package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/usecase"
)

// SyncAnimesArgsはフェーズ2のフル・リコンシリエーションバッチジョブの引数型。
// ペイロードは持たない。ジョブはworks / episodesテーブル全体をリコンサイルするため、
// 実行ごとにパラメータ化するものがない。
type SyncAnimesArgs struct{}

// Kindはジョブの種類を返す。
func (SyncAnimesArgs) Kind() string {
	return "sync_animes"
}

// InsertOptsはジョブ挿入時のデフォルトオプションを返す。リコンサイルは冪等なので、
// 一時的な失敗は数回まで安全に再試行でき、取りこぼしは次回の定期実行でも拾われる。
func (SyncAnimesArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// AnimesSyncerはworks/episodes -> animesのリコンサイルを実行する。
// 実体は *usecase.SyncAnimesUsecase。
type AnimesSyncer interface {
	Execute(ctx context.Context) (*usecase.SyncAnimesResult, error)
}

// SyncAnimesWorkerはフェーズ2のリコンサイルバッチの薄いRiverアダプタ。
type SyncAnimesWorker struct {
	river.WorkerDefaults[SyncAnimesArgs]
	syncer AnimesSyncer
}

// NewSyncAnimesWorkerはSyncAnimesWorkerを生成する。
func NewSyncAnimesWorker(syncer AnimesSyncer) *SyncAnimesWorker {
	return &SyncAnimesWorker{syncer: syncer}
}

// Workはリコンサイルを実行する。件数はUseCase内でログ出力されるため、Workerは
// エラーをそのまま伝搬するだけにとどめ、ジョブ実行ログとリトライはRiverに任せる。
func (w *SyncAnimesWorker) Work(ctx context.Context, job *river.Job[SyncAnimesArgs]) error {
	_, err := w.syncer.Execute(ctx)
	return err
}
