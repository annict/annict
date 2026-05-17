package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
)

// SendSignUpCodeEmailWorker は新規登録確認コードメール送信ワーカーです
type SendSignUpCodeEmailWorker struct {
	river.WorkerDefaults[dispatcher.SendSignUpCodeEmailArgs]
	uc *usecase.SendSignUpCodeEmailUsecase
}

// NewSendSignUpCodeEmailWorker は新しい SendSignUpCodeEmailWorker を作成します
func NewSendSignUpCodeEmailWorker(uc *usecase.SendSignUpCodeEmailUsecase) *SendSignUpCodeEmailWorker {
	return &SendSignUpCodeEmailWorker{uc: uc}
}

// Work は新規登録確認コードメールを送信します
func (w *SendSignUpCodeEmailWorker) Work(ctx context.Context, job *river.Job[dispatcher.SendSignUpCodeEmailArgs]) error {
	return w.uc.Execute(ctx, usecase.SendSignUpCodeEmailInput{
		Email:  job.Args.Email,
		Code:   job.Args.Code,
		Locale: job.Args.Locale,
	})
}
