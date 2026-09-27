package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
)

// SendSignInCodeEmailWorkerはログインコードメール送信ワーカーです
type SendSignInCodeEmailWorker struct {
	river.WorkerDefaults[dispatcher.SendSignInCodeEmailArgs]
	uc *usecase.SendSignInCodeEmailUsecase
}

// NewSendSignInCodeEmailWorkerは新しいSendSignInCodeEmailWorkerを作成します
func NewSendSignInCodeEmailWorker(uc *usecase.SendSignInCodeEmailUsecase) *SendSignInCodeEmailWorker {
	return &SendSignInCodeEmailWorker{uc: uc}
}

// Workはログインコードメールを送信します
func (w *SendSignInCodeEmailWorker) Work(ctx context.Context, job *river.Job[dispatcher.SendSignInCodeEmailArgs]) error {
	return w.uc.Execute(ctx, usecase.SendSignInCodeEmailInput{
		Email:  job.Args.Email,
		Code:   job.Args.Code,
		Locale: job.Args.Locale,
	})
}
