package usecase

import (
	"context"
	"fmt"
	"log/slog"
)

// SignInCodeEmailSender はログインコードメールの送信を行うインターフェース
type SignInCodeEmailSender interface {
	Send(ctx context.Context, to, code, locale string) error
}

// SendSignInCodeEmailUsecase はログインコードメールの送信ユースケース
type SendSignInCodeEmailUsecase struct {
	sender SignInCodeEmailSender
}

// NewSendSignInCodeEmailUsecase は SendSignInCodeEmailUsecase を生成する
func NewSendSignInCodeEmailUsecase(sender SignInCodeEmailSender) *SendSignInCodeEmailUsecase {
	return &SendSignInCodeEmailUsecase{sender: sender}
}

// SendSignInCodeEmailInput はログインコードメール送信の入力パラメータ
type SendSignInCodeEmailInput struct {
	Email  string
	Code   string
	Locale string
}

// Execute はログインコードメールを送信する
func (uc *SendSignInCodeEmailUsecase) Execute(ctx context.Context, input SendSignInCodeEmailInput) error {
	if input.Email == "" {
		return fmt.Errorf("メールアドレスが空です")
	}

	if err := uc.sender.Send(ctx, input.Email, input.Code, input.Locale); err != nil {
		slog.ErrorContext(ctx, "ログインコードメール送信に失敗しました",
			"email", input.Email,
			"error", err,
		)
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	slog.InfoContext(ctx, "ログインコードメールを送信しました",
		"email", input.Email,
	)

	return nil
}
