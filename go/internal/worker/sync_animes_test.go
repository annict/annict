package worker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/worker"
)

// fakeAnimesSyncerはワーカーアダプタのテスト用のAnimesSyncerスタブ。呼び出しを
// 記録し、設定された結果 / エラーを返す。
type fakeAnimesSyncer struct {
	called bool
	result *usecase.SyncAnimesResult
	err    error
}

func (s *fakeAnimesSyncer) Execute(_ context.Context) (*usecase.SyncAnimesResult, error) {
	s.called = true
	return s.result, s.err
}

func TestSyncAnimesWorker_Work_CallsSyncer(t *testing.T) {
	t.Parallel()

	syncer := &fakeAnimesSyncer{result: &usecase.SyncAnimesResult{}}
	w := worker.NewSyncAnimesWorker(syncer)

	job := &river.Job[worker.SyncAnimesArgs]{Args: worker.SyncAnimesArgs{}}
	if err := w.Work(context.Background(), job); err != nil {
		t.Fatalf("Work()のエラー = %v", err)
	}
	if !syncer.called {
		t.Error("syncer.Execute()が呼ばれなかった")
	}
}

func TestSyncAnimesWorker_Work_PropagatesError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("sync boom")
	syncer := &fakeAnimesSyncer{err: wantErr}
	w := worker.NewSyncAnimesWorker(syncer)

	job := &river.Job[worker.SyncAnimesArgs]{Args: worker.SyncAnimesArgs{}}
	if err := w.Work(context.Background(), job); !errors.Is(err, wantErr) {
		t.Fatalf("Work()のエラー = %v、期待値 = %v", err, wantErr)
	}
}

func TestSyncAnimesArgs_Kind(t *testing.T) {
	t.Parallel()

	// kind文字列は永続化されるジョブ識別子。リネームで予定済みジョブが孤立する
	// のを検出できるよう固定する。
	if got := (worker.SyncAnimesArgs{}).Kind(); got != "sync_animes" {
		t.Errorf("Kind() = %q、期待値 = sync_animes", got)
	}
}
