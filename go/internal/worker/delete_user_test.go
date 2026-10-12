package worker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
	"github.com/annict/annict/go/internal/worker"
)

type userDeleterStub struct {
	input  usecase.DeleteUserInput
	result *usecase.DeleteUserResult
	err    error
}

func (s *userDeleterStub) Execute(_ context.Context, input usecase.DeleteUserInput) (*usecase.DeleteUserResult, error) {
	s.input = input
	return s.result, s.err
}

func TestDeleteUserWorker_Work(t *testing.T) {
	t.Parallel()

	failure := errors.New("削除に失敗")
	notDeleted := fmt.Errorf("%w (ユーザーID: 1)", usecase.ErrUserNotDeleted)

	tests := []struct {
		name       string
		result     *usecase.DeleteUserResult
		err        error
		assertWork func(t *testing.T, err error)
	}{
		{
			name:   "正常系: すべての段階を終えたらnilを返す",
			result: &usecase.DeleteUserResult{Completed: true},
			assertWork: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("Work()のエラー = %v、期待値 = nil", err)
				}
			},
		},
		{
			name:   "正常系: 実行時間の上限で中断したらsnoozeする",
			result: &usecase.DeleteUserResult{Completed: false},
			assertWork: func(t *testing.T, err error) {
				var snoozeErr *river.JobSnoozeError
				if !errors.As(err, &snoozeErr) {
					t.Fatalf("Work()のエラー = %v、期待値 = JobSnoozeError", err)
				}
				if snoozeErr.Duration != worker.DeleteUserSnoozeDelay {
					t.Errorf("snoozeの間隔 = %v、期待値 = %v", snoozeErr.Duration, worker.DeleteUserSnoozeDelay)
				}
			},
		},
		{
			name: "異常系: 論理削除されていないユーザーはジョブを取り消す",
			err:  notDeleted,
			assertWork: func(t *testing.T, err error) {
				var cancelErr *river.JobCancelError
				if !errors.As(err, &cancelErr) {
					t.Fatalf("Work()のエラー = %v、期待値 = JobCancelError", err)
				}
				if !errors.Is(err, usecase.ErrUserNotDeleted) {
					t.Errorf("Work()のエラー = %v、ErrUserNotDeletedを含むべき", err)
				}
			},
		},
		{
			name: "異常系: そのほかのエラーはそのまま返して再試行させる",
			err:  failure,
			assertWork: func(t *testing.T, err error) {
				if !errors.Is(err, failure) {
					t.Errorf("Work()のエラー = %v、期待値 = %v", err, failure)
				}
				var cancelErr *river.JobCancelError
				if errors.As(err, &cancelErr) {
					t.Error("再試行すべきエラーでジョブが取り消された")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			deleter := &userDeleterStub{result: tt.result, err: tt.err}
			w := worker.NewDeleteUserWorker(deleter)
			job := &river.Job[dispatcher.DeleteUserArgs]{
				JobRow: &rivertype.JobRow{Attempt: 1, MaxAttempts: 10},
				Args:   dispatcher.DeleteUserArgs{UserID: 42},
			}

			err := w.Work(context.Background(), job)
			tt.assertWork(t, err)
			if deleter.input.UserID != model.UserID(42) {
				t.Errorf("UserID = %v、期待値 = 42", deleter.input.UserID)
			}
		})
	}
}

func TestDeleteUserWorker_Work_Canceled(t *testing.T) {
	t.Parallel()

	// ワーカーの停止で中断した場合は、待たずに再投入する。
	deleter := &userDeleterStub{result: &usecase.DeleteUserResult{Completed: false}}
	w := worker.NewDeleteUserWorker(deleter)
	job := &river.Job[dispatcher.DeleteUserArgs]{
		JobRow: &rivertype.JobRow{Attempt: 1, MaxAttempts: 10},
		Args:   dispatcher.DeleteUserArgs{UserID: 42},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := w.Work(ctx, job)
	var snoozeErr *river.JobSnoozeError
	if !errors.As(err, &snoozeErr) {
		t.Fatalf("Work()のエラー = %v、期待値 = JobSnoozeError", err)
	}
	if snoozeErr.Duration != 0 {
		t.Errorf("snoozeの間隔 = %v、期待値 = 0", snoozeErr.Duration)
	}
}

func TestDeleteUserWorker_Work_FinalAttempt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		attempt int
		want    bool
	}{
		{name: "最後の試行より前はfalse", attempt: 9, want: false},
		{name: "最後の試行はtrue", attempt: 10, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			deleter := &userDeleterStub{result: &usecase.DeleteUserResult{Completed: true}}
			w := worker.NewDeleteUserWorker(deleter)
			job := &river.Job[dispatcher.DeleteUserArgs]{
				JobRow: &rivertype.JobRow{Attempt: tt.attempt, MaxAttempts: 10},
				Args:   dispatcher.DeleteUserArgs{UserID: 42},
			}

			if err := w.Work(context.Background(), job); err != nil {
				t.Fatalf("Work()のエラー = %v", err)
			}
			if deleter.input.FinalAttempt != tt.want {
				t.Errorf("FinalAttempt = %v、期待値 = %v", deleter.input.FinalAttempt, tt.want)
			}
		})
	}
}

func TestDeleteUserWorker_Timeout(t *testing.T) {
	t.Parallel()

	w := worker.NewDeleteUserWorker(&userDeleterStub{})

	// UseCaseの上限時間で打ち切った実行がタイムアウトのエラーにならないよう、上限時間より長くする。
	got := w.Timeout(&river.Job[dispatcher.DeleteUserArgs]{})
	if got <= usecase.DeleteUserTimeLimit {
		t.Errorf("Timeout() = %v、UseCaseの上限時間 %v より長くなければならない", got, usecase.DeleteUserTimeLimit)
	}
	if got <= river.JobTimeoutDefault {
		t.Errorf("Timeout() = %v、Riverの既定値 %v より長くなければならない", got, river.JobTimeoutDefault)
	}
}
