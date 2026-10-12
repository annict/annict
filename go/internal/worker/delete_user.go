package worker

import (
	"context"
	"errors"
	"time"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// DeleteUserTimeoutはDeleteUserWorkerの1回の試行の実行上限。Riverの既定のジョブタイムアウト
// (1分) を上書きし、UseCaseの上限時間に達したあとの最後の1単位のぶんの余裕を足す。
const DeleteUserTimeout = usecase.DeleteUserTimeLimit + 5*time.Minute

// DeleteUserSnoozeDelayは実行時間の上限で中断した削除ジョブを再開するまでの間隔。
// 記録の多いユーザーの削除が、ほかのジョブやリクエストのDBアクセスと間を空けて進むようにする。
const DeleteUserSnoozeDelay = time.Minute

// UserDeleterはユーザーの削除を実行する。実体は *usecase.DeleteUserUsecase。
type UserDeleter interface {
	Execute(ctx context.Context, input usecase.DeleteUserInput) (*usecase.DeleteUserResult, error)
}

// DeleteUserWorkerは退会・利用停止したユーザーのデータを削除するワーカー。
type DeleteUserWorker struct {
	river.WorkerDefaults[dispatcher.DeleteUserArgs]
	deleter UserDeleter
}

// NewDeleteUserWorkerは新しいDeleteUserWorkerを作成する。
func NewDeleteUserWorker(deleter UserDeleter) *DeleteUserWorker {
	return &DeleteUserWorker{deleter: deleter}
}

// Timeoutはジョブの実行上限を返す。
func (w *DeleteUserWorker) Timeout(*river.Job[dispatcher.DeleteUserArgs]) time.Duration {
	return DeleteUserTimeout
}

// Workはユーザーの削除を進める。UseCaseの結果をRiverの制御に読み替えるだけにとどめ、
// 実行時間の上限やワーカーの停止で中断した場合はsnoozeで続きを再投入する (snoozeは試行回数を消費しない)。
// 論理削除されていないユーザーは再試行しても結果が変わらないため、ジョブを取り消す。
func (w *DeleteUserWorker) Work(ctx context.Context, job *river.Job[dispatcher.DeleteUserArgs]) error {
	result, err := w.deleter.Execute(ctx, usecase.DeleteUserInput{
		UserID:       model.UserID(job.Args.UserID),
		FinalAttempt: job.Attempt >= job.MaxAttempts,
	})
	if errors.Is(err, usecase.ErrUserNotDeleted) {
		return river.JobCancel(err)
	}
	if err != nil {
		return err
	}
	if !result.Completed {
		// ワーカーの停止で中断した場合は、試行回数を消費せず、次に起動したクライアントが
		// すぐ続きを進められるよう待たずに再投入する。
		if errors.Is(ctx.Err(), context.Canceled) {
			return river.JobSnooze(0)
		}
		return river.JobSnooze(DeleteUserSnoozeDelay)
	}
	return nil
}
