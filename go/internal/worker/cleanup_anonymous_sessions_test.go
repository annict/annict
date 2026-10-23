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

type anonymousSessionCleanerStub struct {
	called bool
	err    error
}

func (s *anonymousSessionCleanerStub) Execute(_ context.Context) error {
	s.called = true
	return s.err
}

func TestCleanupAnonymousSessionsWorker_Work(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("cleanup failed")
	tests := []struct {
		name    string
		wantErr error
	}{
		{name: "正常系: cleanerの成功をそのまま返す"},
		{name: "異常系: cleanerのエラーをそのまま返す", wantErr: wantErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := &anonymousSessionCleanerStub{err: tt.wantErr}
			w := worker.NewCleanupAnonymousSessionsWorker(cleaner)
			job := &river.Job[dispatcher.CleanupAnonymousSessionsArgs]{
				Args: dispatcher.CleanupAnonymousSessionsArgs{},
			}

			err := w.Work(context.Background(), job)
			if !cleaner.called {
				t.Error("Execute()が呼ばれていません")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Work()のエラー = %v、期待値 = %v", err, tt.wantErr)
			}
		})
	}
}

func TestCleanupAnonymousSessionsWorker_Timeout(t *testing.T) {
	t.Parallel()

	w := worker.NewCleanupAnonymousSessionsWorker(&anonymousSessionCleanerStub{})

	// UseCaseの上限時間で打ち切った実行がタイムアウトのエラーにならないよう、上限時間より
	// 長くする。前の実行と次の実行が重ならないよう、実行間隔 (1時間) より短いことも確認する。
	got := w.Timeout(&river.Job[dispatcher.CleanupAnonymousSessionsArgs]{})
	if got != 10*time.Minute {
		t.Errorf("Timeout() = %v、期待値 = 10m", got)
	}
	if got <= usecase.CleanupAnonymousSessionsTimeLimit {
		t.Errorf("Timeout() = %v、UseCaseの上限時間 %v より長くなければならない", got, usecase.CleanupAnonymousSessionsTimeLimit)
	}
	if got <= river.JobTimeoutDefault {
		t.Errorf("Timeout() = %v、Riverの既定値 %v より長くなければならない", got, river.JobTimeoutDefault)
	}
	if got >= time.Hour {
		t.Errorf("Timeout() = %v、実行間隔 (1時間) より短くなければならない", got)
	}
}
