package sign_up

import (
	"context"
	"net/mail"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreateValidator は新規登録フォームのバリデーションを行う
type CreateValidator struct{}

// NewCreateValidator は CreateValidator を生成する
func NewCreateValidator() *CreateValidator {
	return &CreateValidator{}
}

// CreateValidatorInput はバリデーションの入力パラメータ
type CreateValidatorInput struct {
	Email string
}

// CreateValidatorResult はバリデーションの結果
type CreateValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateValidator) Validate(ctx context.Context, input CreateValidatorInput) *CreateValidatorResult {
	formErrors := &session.FormErrors{}

	// メールアドレスの必須チェック
	if strings.TrimSpace(input.Email) == "" {
		formErrors.AddFieldError("email", i18n.T(ctx, "sign_up_email_required"))
		return &CreateValidatorResult{FormErrors: formErrors}
	}

	// メールアドレスの形式チェック
	if _, err := mail.ParseAddress(input.Email); err != nil {
		formErrors.AddFieldError("email", i18n.T(ctx, "sign_up_email_invalid"))
		return &CreateValidatorResult{FormErrors: formErrors}
	}

	return &CreateValidatorResult{}
}
