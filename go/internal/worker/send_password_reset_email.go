package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
)

// SendPasswordResetEmailWorkerはパスワードリセットメール送信ワーカーです
type SendPasswordResetEmailWorker struct {
	river.WorkerDefaults[dispatcher.SendPasswordResetEmailArgs]
	uc *usecase.SendPasswordResetEmailUsecase
}

// NewSendPasswordResetEmailWorkerは新しいSendPasswordResetEmailWorkerを作成します
func NewSendPasswordResetEmailWorker(uc *usecase.SendPasswordResetEmailUsecase) *SendPasswordResetEmailWorker {
	return &SendPasswordResetEmailWorker{uc: uc}
}

// Workはパスワードリセットメールを送信します
func (w *SendPasswordResetEmailWorker) Work(ctx context.Context, job *river.Job[dispatcher.SendPasswordResetEmailArgs]) error {
	return w.uc.Execute(ctx, usecase.SendPasswordResetEmailInput{
		Email:    job.Args.Email,
		ResetURL: job.Args.ResetURL,
		Locale:   job.Args.Locale,
	})
}
