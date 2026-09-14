package usecase

import (
	"context"
	"fmt"
	"log/slog"
)

// PasswordResetEmailSenderはパスワードリセットメールの送信を行うインターフェース
type PasswordResetEmailSender interface {
	Send(ctx context.Context, to, resetURL, locale string) error
}

// SendPasswordResetEmailUsecaseはパスワードリセットメールの送信ユースケース
type SendPasswordResetEmailUsecase struct {
	sender PasswordResetEmailSender
}

// NewSendPasswordResetEmailUsecaseはSendPasswordResetEmailUsecaseを生成する
func NewSendPasswordResetEmailUsecase(sender PasswordResetEmailSender) *SendPasswordResetEmailUsecase {
	return &SendPasswordResetEmailUsecase{sender: sender}
}

// SendPasswordResetEmailInputはパスワードリセットメール送信の入力パラメータ
type SendPasswordResetEmailInput struct {
	Email    string
	ResetURL string
	Locale   string
}

// Executeはパスワードリセットメールを送信する
func (uc *SendPasswordResetEmailUsecase) Execute(ctx context.Context, input SendPasswordResetEmailInput) error {
	if input.Email == "" {
		return fmt.Errorf("メールアドレスが空です")
	}

	if err := uc.sender.Send(ctx, input.Email, input.ResetURL, input.Locale); err != nil {
		slog.ErrorContext(ctx, "パスワードリセットメール送信に失敗しました",
			"email", input.Email,
			"error", err,
		)
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	slog.InfoContext(ctx, "パスワードリセットメールを送信しました",
		"email", input.Email,
	)

	return nil
}
