package usecase

import (
	"context"
	"fmt"
	"log/slog"
)

// SignUpCodeEmailSender は新規登録確認コードメールの送信を行うインターフェース
type SignUpCodeEmailSender interface {
	Send(ctx context.Context, to, code, locale string) error
}

// SendSignUpCodeEmailUsecase は新規登録確認コードメールの送信ユースケース
type SendSignUpCodeEmailUsecase struct {
	sender SignUpCodeEmailSender
}

// NewSendSignUpCodeEmailUsecase は SendSignUpCodeEmailUsecase を生成する
func NewSendSignUpCodeEmailUsecase(sender SignUpCodeEmailSender) *SendSignUpCodeEmailUsecase {
	return &SendSignUpCodeEmailUsecase{sender: sender}
}

// SendSignUpCodeEmailInput は新規登録確認コードメール送信の入力パラメータ
type SendSignUpCodeEmailInput struct {
	Email  string
	Code   string
	Locale string
}

// Execute は新規登録確認コードメールを送信する
func (uc *SendSignUpCodeEmailUsecase) Execute(ctx context.Context, input SendSignUpCodeEmailInput) error {
	if input.Email == "" {
		return fmt.Errorf("メールアドレスが空です")
	}

	if err := uc.sender.Send(ctx, input.Email, input.Code, input.Locale); err != nil {
		slog.ErrorContext(ctx, "新規登録確認コードメール送信に失敗しました",
			"email", input.Email,
			"error", err,
		)
		return fmt.Errorf("メール送信に失敗: %w", err)
	}

	slog.InfoContext(ctx, "新規登録確認コードメールを送信しました",
		"email", input.Email,
	)

	return nil
}
