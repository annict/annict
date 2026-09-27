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

type expiredSessionCleanerStub struct {
	called bool
	err    error
}

func (s *expiredSessionCleanerStub) Execute(_ context.Context) error {
	s.called = true
	return s.err
}

func TestCleanupExpiredSessionsWorker_Work(t *testing.T) {
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
			cleaner := &expiredSessionCleanerStub{err: tt.wantErr}
			w := worker.NewCleanupExpiredSessionsWorker(cleaner)
			job := &river.Job[dispatcher.CleanupExpiredSessionsArgs]{
				Args: dispatcher.CleanupExpiredSessionsArgs{},
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

func TestCleanupExpiredSessionsWorker_Timeout(t *testing.T) {
	t.Parallel()

	w := worker.NewCleanupExpiredSessionsWorker(&expiredSessionCleanerStub{})

	// UseCaseの上限時間で打ち切った実行がタイムアウトのエラーにならないよう、上限時間より
	// 長くする。前の実行と次の実行が重ならないよう、実行間隔 (1時間) より短いことも確認する。
	got := w.Timeout(&river.Job[dispatcher.CleanupExpiredSessionsArgs]{})
	if got != 15*time.Minute {
		t.Errorf("Timeout() = %v、期待値 = 15m", got)
	}
	if got <= usecase.CleanupExpiredSessionsTimeLimit {
		t.Errorf("Timeout() = %v、UseCaseの上限時間 %v より長くなければならない", got, usecase.CleanupExpiredSessionsTimeLimit)
	}
	if got <= river.JobTimeoutDefault {
		t.Errorf("Timeout() = %v、Riverの既定値 %v より長くなければならない", got, river.JobTimeoutDefault)
	}
	if got >= time.Hour {
		t.Errorf("Timeout() = %v、実行間隔 (1時間) より短くなければならない", got)
	}
}
