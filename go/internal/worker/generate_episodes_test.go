package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/worker"
)

// fakeEpisodesGeneratorはワーカーアダプタのテスト用のEpisodesGeneratorスタブ。受け取った
// 入力を記録し、設定されたエラーを返す。
type fakeEpisodesGenerator struct {
	input *usecase.GenerateEpisodesInput
	err   error
}

func (g *fakeEpisodesGenerator) Execute(_ context.Context, input usecase.GenerateEpisodesInput) (*usecase.GenerateEpisodesOutput, error) {
	g.input = &input
	if g.err != nil {
		return nil, g.err
	}
	return &usecase.GenerateEpisodesOutput{}, nil
}

func TestGenerateEpisodesWorker_Work_PassesCurrentTime(t *testing.T) {
	t.Parallel()

	generator := &fakeEpisodesGenerator{}
	w := worker.NewGenerateEpisodesWorker(generator)

	before := time.Now()
	job := &river.Job[dispatcher.GenerateEpisodesArgs]{Args: dispatcher.GenerateEpisodesArgs{}}
	if err := w.Work(context.Background(), job); err != nil {
		t.Fatalf("Work()のエラー = %v", err)
	}
	if generator.input == nil {
		t.Fatal("generator.Execute()が呼ばれなかった")
	}
	if generator.input.Now.Before(before) || generator.input.Now.After(time.Now()) {
		t.Errorf("Now = %v、期待値 = 実行した時刻", generator.input.Now)
	}
}

func TestGenerateEpisodesWorker_Work_PropagatesError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("報告メールの送信に失敗")
	w := worker.NewGenerateEpisodesWorker(&fakeEpisodesGenerator{err: wantErr})

	job := &river.Job[dispatcher.GenerateEpisodesArgs]{Args: dispatcher.GenerateEpisodesArgs{}}
	if err := w.Work(context.Background(), job); !errors.Is(err, wantErr) {
		t.Fatalf("Work()のエラー = %v、期待値 = %v", err, wantErr)
	}
}

func TestGenerateEpisodesWorker_Timeout(t *testing.T) {
	t.Parallel()

	w := worker.NewGenerateEpisodesWorker(&fakeEpisodesGenerator{})

	got := w.Timeout(&river.Job[dispatcher.GenerateEpisodesArgs]{})
	if got <= river.JobTimeoutDefault {
		t.Errorf("Timeout() = %v、Riverの既定値 %v より長くなければならない", got, river.JobTimeoutDefault)
	}
}
