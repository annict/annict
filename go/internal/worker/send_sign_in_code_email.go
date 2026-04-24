package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
)

// SendSignInCodeEmailWorker はログインコードメール送信ワーカーです
type SendSignInCodeEmailWorker struct {
	river.WorkerDefaults[dispatcher.SendSignInCodeEmailArgs]
	uc *usecase.SendSignInCodeEmailUsecase
}

// NewSendSignInCodeEmailWorker は新しい SendSignInCodeEmailWorker を作成します
func NewSendSignInCodeEmailWorker(uc *usecase.SendSignInCodeEmailUsecase) *SendSignInCodeEmailWorker {
	return &SendSignInCodeEmailWorker{uc: uc}
}

// Work はログインコードメールを送信します
func (w *SendSignInCodeEmailWorker) Work(ctx context.Context, job *river.Job[dispatcher.SendSignInCodeEmailArgs]) error {
	return w.uc.Execute(ctx, usecase.SendSignInCodeEmailInput{
		Email:  job.Args.Email,
		Code:   job.Args.Code,
		Locale: job.Args.Locale,
	})
}
