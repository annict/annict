package validator

import (
	"context"
	"net/mail"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreateSignUpValidator は新規登録フォームのバリデーションを行う
type CreateSignUpValidator struct{}

// NewCreateSignUpValidator は CreateSignUpValidator を生成する
func NewCreateSignUpValidator() *CreateSignUpValidator {
	return &CreateSignUpValidator{}
}

// CreateSignUpValidatorInput はバリデーションの入力パラメータ
type CreateSignUpValidatorInput struct {
	Email string
}

// CreateSignUpValidatorResult はバリデーションの結果
type CreateSignUpValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はフォームの形式バリデーションを行う
func (v *CreateSignUpValidator) Validate(ctx context.Context, input CreateSignUpValidatorInput) *CreateSignUpValidatorResult {
	formErrors := &session.FormErrors{}

	// メールアドレスの必須チェック
	if strings.TrimSpace(input.Email) == "" {
		formErrors.AddFieldError("email", i18n.T(ctx, "sign_up_email_required"))
		return &CreateSignUpValidatorResult{FormErrors: formErrors}
	}

	// メールアドレスの形式チェック
	if _, err := mail.ParseAddress(input.Email); err != nil {
		formErrors.AddFieldError("email", i18n.T(ctx, "sign_up_email_invalid"))
		return &CreateSignUpValidatorResult{FormErrors: formErrors}
	}

	return &CreateSignUpValidatorResult{}
}
