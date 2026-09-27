package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestSyncAnimesWorker_Timeout(t *testing.T) {
	t.Parallel()

	w := worker.NewSyncAnimesWorker(&fakeAnimesSyncer{})

	// Riverの既定のタイムアウト (1分) では全件走査が打ち切られるため上書きしている。
	// 実行間隔 (1時間) より短いことも確認する。
	got := w.Timeout(&river.Job[worker.SyncAnimesArgs]{})
	if got != 30*time.Minute {
		t.Errorf("Timeout() = %v、期待値 = 30m", got)
	}
	if got <= river.JobTimeoutDefault {
		t.Errorf("Timeout() = %v、Riverの既定値 %v より長くなければならない", got, river.JobTimeoutDefault)
	}
	if got >= time.Hour {
		t.Errorf("Timeout() = %v、実行間隔 (1時間) より短くなければならない", got)
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
