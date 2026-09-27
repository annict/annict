package worker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
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
