package usecase

import (
	"context"
	"fmt"
	"log/slog"
)

// AccountDeletionEmailSenderは退会完了メールの送信を行うインターフェース
type AccountDeletionEmailSender interface {
	Send(ctx context.Context, to, locale string, isStripeSupporter bool) error
}

// SendAccountDeletionEmailUsecaseは退会完了メールの送信ユースケース
type SendAccountDeletionEmailUsecase struct {
	sender AccountDeletionEmailSender
}

// NewSendAccountDeletionEmailUsecaseはSendAccountDeletionEmailUsecaseを生成する
func NewSendAccountDeletionEmailUsecase(sender AccountDeletionEmailSender) *SendAccountDeletionEmailUsecase {
	return &SendAccountDeletionEmailUsecase{sender: sender}
}

// SendAccountDeletionEmailInputは退会完了メール送信の入力パラメータ
type SendAccountDeletionEmailInput struct {
	Email             string
	Locale            string
	IsStripeSupporter bool
}

// Executeは退会完了メールを送信する。退会したユーザーのメールアドレスはログに残さない。
func (uc *SendAccountDeletionEmailUsecase) Execute(ctx context.Context, input SendAccountDeletionEmailInput) error {
	if input.Email == "" {
		return fmt.Errorf("メールアドレスが空です")
	}

	if err := uc.sender.Send(ctx, input.Email, input.Locale, input.IsStripeSupporter); err != nil {
		slog.ErrorContext(ctx, "退会完了メールの送信に失敗しました", "error", err)
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	slog.InfoContext(ctx, "退会完了メールを送信しました")

	return nil
}
