package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/annict/annict/go/internal/model"
)

// TwoFactorAuthNotificationEmailSenderは、二要素認証の設定の変更を知らせるメールの送信を行うインターフェース
type TwoFactorAuthNotificationEmailSender interface {
	Send(ctx context.Context, to, locale string, event model.TwoFactorAuthNotificationEvent) error
}

// SendTwoFactorAuthNotificationEmailUsecaseは、二要素認証の設定の変更を知らせるメールの送信ユースケース
type SendTwoFactorAuthNotificationEmailUsecase struct {
	sender TwoFactorAuthNotificationEmailSender
}

// NewSendTwoFactorAuthNotificationEmailUsecaseはSendTwoFactorAuthNotificationEmailUsecaseを生成する
func NewSendTwoFactorAuthNotificationEmailUsecase(sender TwoFactorAuthNotificationEmailSender) *SendTwoFactorAuthNotificationEmailUsecase {
	return &SendTwoFactorAuthNotificationEmailUsecase{sender: sender}
}

// SendTwoFactorAuthNotificationEmailInputは通知メールの送信の入力パラメータ
type SendTwoFactorAuthNotificationEmailInput struct {
	Email  string
	Locale string
	Event  model.TwoFactorAuthNotificationEvent
}

// Executeは通知メールを送信する。送信先のメールアドレスはログに残さない。
func (uc *SendTwoFactorAuthNotificationEmailUsecase) Execute(ctx context.Context, input SendTwoFactorAuthNotificationEmailInput) error {
	if input.Email == "" {
		return fmt.Errorf("メールアドレスが空です")
	}
	if !input.Event.IsValid() {
		return fmt.Errorf("未知の二要素認証の通知の種類です: %q", input.Event)
	}

	if err := uc.sender.Send(ctx, input.Email, input.Locale, input.Event); err != nil {
		slog.ErrorContext(ctx, "二要素認証の通知メールの送信に失敗しました", "error", err, "event", input.Event)
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	slog.InfoContext(ctx, "二要素認証の通知メールを送信しました", "event", input.Event)

	return nil
}
