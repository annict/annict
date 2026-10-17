package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/usecase"
)

// SendAccountDeletionEmailWorkerは退会完了メール送信ワーカーです
type SendAccountDeletionEmailWorker struct {
	river.WorkerDefaults[dispatcher.SendAccountDeletionEmailArgs]
	uc *usecase.SendAccountDeletionEmailUsecase
}

// NewSendAccountDeletionEmailWorkerは新しいSendAccountDeletionEmailWorkerを作成します
func NewSendAccountDeletionEmailWorker(uc *usecase.SendAccountDeletionEmailUsecase) *SendAccountDeletionEmailWorker {
	return &SendAccountDeletionEmailWorker{uc: uc}
}

// Workは退会完了メールを送信します
func (w *SendAccountDeletionEmailWorker) Work(ctx context.Context, job *river.Job[dispatcher.SendAccountDeletionEmailArgs]) error {
	return w.uc.Execute(ctx, usecase.SendAccountDeletionEmailInput{
		Email:             job.Args.Email,
		Locale:            job.Args.Locale,
		IsStripeSupporter: job.Args.IsStripeSupporter,
	})
}
