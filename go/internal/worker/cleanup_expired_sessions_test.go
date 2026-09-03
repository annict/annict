package worker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/riverqueue/river"

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

func TestCleanupExpiredSessionsArgs_Kind(t *testing.T) {
	t.Parallel()

	if got, want := (worker.CleanupExpiredSessionsArgs{}).Kind(), "cleanup_expired_sessions"; got != want {
		t.Errorf("Kind() = %q, want %q", got, want)
	}
}

func TestCleanupExpiredSessionsWorker_Work(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("cleanup failed")
	tests := []struct {
		name    string
		wantErr error
	}{
		{name: "正常系: cleaner の成功をそのまま返す"},
		{name: "異常系: cleaner のエラーをそのまま返す", wantErr: wantErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaner := &expiredSessionCleanerStub{err: tt.wantErr}
			w := worker.NewCleanupExpiredSessionsWorker(cleaner)
			job := &river.Job[worker.CleanupExpiredSessionsArgs]{
				Args: worker.CleanupExpiredSessionsArgs{},
			}

			err := w.Work(context.Background(), job)
			if !cleaner.called {
				t.Error("Execute() が呼ばれていません")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Work() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
