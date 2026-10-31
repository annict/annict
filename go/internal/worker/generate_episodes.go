package worker

import (
	"context"
	"time"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
)

// GenerateEpisodesTimeoutはGenerateEpisodesWorkerの1回の試行の実行上限。Riverの既定の
// ジョブタイムアウト (1分) では、前後7日の放送枠を1件ずつ処理し、しょぼいカレンダーへ問い
// 合わせるまでに足りないことがあるため上書きする。
const GenerateEpisodesTimeout = 10 * time.Minute

// EpisodesGeneratorはエピソードの自動生成を実行する。
// 実体は *usecase.GenerateEpisodesUsecase。
type EpisodesGenerator interface {
	Execute(ctx context.Context, input usecase.GenerateEpisodesInput) (*usecase.GenerateEpisodesOutput, error)
}

// GenerateEpisodesWorkerはエピソードの自動生成の薄いRiverアダプタ。
type GenerateEpisodesWorker struct {
	river.WorkerDefaults[dispatcher.GenerateEpisodesArgs]
	generator EpisodesGenerator
}

// NewGenerateEpisodesWorkerはGenerateEpisodesWorkerを生成する。
func NewGenerateEpisodesWorker(generator EpisodesGenerator) *GenerateEpisodesWorker {
	return &GenerateEpisodesWorker{generator: generator}
}

// Timeoutはジョブの実行上限を返す。
func (w *GenerateEpisodesWorker) Timeout(*river.Job[dispatcher.GenerateEpisodesArgs]) time.Duration {
	return GenerateEpisodesTimeout
}

// Workは現在時刻を実行時刻として自動生成を実行する。結果はUseCase内でログと報告メールに
// 残すため、Workerはエラーをそのまま伝搬するだけにとどめる。
func (w *GenerateEpisodesWorker) Work(ctx context.Context, job *river.Job[dispatcher.GenerateEpisodesArgs]) error {
	_, err := w.generator.Execute(ctx, usecase.GenerateEpisodesInput{Now: time.Now()})
	return err
}
