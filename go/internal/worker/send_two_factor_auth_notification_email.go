package worker

import (
	"context"

	"github.com/riverqueue/river"

	"github.com/annict/annict/go/internal/dispatcher"
	"github.com/annict/annict/go/internal/model"
	"github.com/annict/annict/go/internal/usecase"
)

// SendTwoFactorAuthNotificationEmailWorkerは、二要素認証の設定の変更を知らせるメールの送信ワーカーです
type SendTwoFactorAuthNotificationEmailWorker struct {
	river.WorkerDefaults[dispatcher.SendTwoFactorAuthNotificationEmailArgs]
	uc *usecase.SendTwoFactorAuthNotificationEmailUsecase
}

// NewSendTwoFactorAuthNotificationEmailWorkerは新しいSendTwoFactorAuthNotificationEmailWorkerを作成します
func NewSendTwoFactorAuthNotificationEmailWorker(uc *usecase.SendTwoFactorAuthNotificationEmailUsecase) *SendTwoFactorAuthNotificationEmailWorker {
	return &SendTwoFactorAuthNotificationEmailWorker{uc: uc}
}

// Workは二要素認証の設定の変更を知らせるメールを送信します
func (w *SendTwoFactorAuthNotificationEmailWorker) Work(ctx context.Context, job *river.Job[dispatcher.SendTwoFactorAuthNotificationEmailArgs]) error {
	return w.uc.Execute(ctx, usecase.SendTwoFactorAuthNotificationEmailInput{
		Email:  job.Args.Email,
		Locale: job.Args.Locale,
		Event:  model.TwoFactorAuthNotificationEvent(job.Args.Event),
	})
}
