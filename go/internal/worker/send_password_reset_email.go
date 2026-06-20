package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
)

// SendPasswordResetEmailWorker はパスワードリセットメール送信ワーカーです
type SendPasswordResetEmailWorker struct {
	river.WorkerDefaults[dispatcher.SendPasswordResetEmailArgs]
	uc *usecase.SendPasswordResetEmailUsecase
}

// NewSendPasswordResetEmailWorker は新しい SendPasswordResetEmailWorker を作成します
func NewSendPasswordResetEmailWorker(uc *usecase.SendPasswordResetEmailUsecase) *SendPasswordResetEmailWorker {
	return &SendPasswordResetEmailWorker{uc: uc}
}

// Work はパスワードリセットメールを送信します
func (w *SendPasswordResetEmailWorker) Work(ctx context.Context, job *river.Job[dispatcher.SendPasswordResetEmailArgs]) error {
	return w.uc.Execute(ctx, usecase.SendPasswordResetEmailInput{
		Email:    job.Args.Email,
		ResetURL: job.Args.ResetURL,
		Locale:   job.Args.Locale,
	})
}
