package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/usecase"
)

// SyncAnimesArgs is the argument type for the phase 2 full-reconciliation batch
// job. It carries no payload: the job reconciles the whole works / episodes tables,
// so there is nothing to parameterize per run.
//
// [Ja] SyncAnimesArgs はフェーズ 2 のフル・リコンシリエーションバッチジョブの引数型。
// ペイロードは持たない。ジョブは works / episodes テーブル全体をリコンサイルするため、
// 実行ごとにパラメータ化するものがない。
type SyncAnimesArgs struct{}

// Kind returns the job kind.
//
// [Ja] Kind はジョブの種類を返す。
func (SyncAnimesArgs) Kind() string {
	return "sync_animes"
}

// InsertOpts returns the default insert options. The reconciliation is idempotent,
// so a transient failure can be safely retried a few times; whatever it misses is
// also caught by the next periodic run.
//
// [Ja] InsertOpts はジョブ挿入時のデフォルトオプションを返す。リコンサイルは冪等なので、
// 一時的な失敗は数回まで安全に再試行でき、取りこぼしは次回の定期実行でも拾われる。
func (SyncAnimesArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// AnimesSyncer runs the works/episodes -> animes reconciliation. Implemented by
// *usecase.SyncAnimesUsecase.
//
// [Ja] AnimesSyncer は works/episodes -> animes のリコンサイルを実行する。
// 実体は *usecase.SyncAnimesUsecase。
type AnimesSyncer interface {
	Execute(ctx context.Context) (*usecase.SyncAnimesResult, error)
}

// SyncAnimesWorker is the thin River adapter for the phase 2 reconciliation batch.
//
// [Ja] SyncAnimesWorker はフェーズ 2 のリコンサイルバッチの薄い River アダプタ。
type SyncAnimesWorker struct {
	river.WorkerDefaults[SyncAnimesArgs]
	syncer AnimesSyncer
}

// NewSyncAnimesWorker constructs a SyncAnimesWorker.
//
// [Ja] NewSyncAnimesWorker は SyncAnimesWorker を生成する。
func NewSyncAnimesWorker(syncer AnimesSyncer) *SyncAnimesWorker {
	return &SyncAnimesWorker{syncer: syncer}
}

// Work runs the reconciliation. The result counts are logged inside the usecase, so
// the worker only propagates the error, leaving job-run logging and retries to
// River.
//
// [Ja] Work はリコンサイルを実行する。件数は UseCase 内でログ出力されるため、Worker は
// エラーをそのまま伝搬するだけにとどめ、ジョブ実行ログとリトライは River に任せる。
func (w *SyncAnimesWorker) Work(ctx context.Context, job *river.Job[SyncAnimesArgs]) error {
	_, err := w.syncer.Execute(ctx)
	return err
}
