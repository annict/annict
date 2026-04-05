package validator

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreatePasswordResetValidator はパスワードリセット申請フォームのバリデーションを行う
type CreatePasswordResetValidator struct{}

// NewCreatePasswordResetValidator は CreatePasswordResetValidator を生成する
func NewCreatePasswordResetValidator() *CreatePasswordResetValidator {
	return &CreatePasswordResetValidator{}
}

// CreatePasswordResetValidatorInput はバリデーションの入力パラメータ
type CreatePasswordResetValidatorInput struct {
	Email string
}

// CreatePasswordResetValidatorResult はバリデーションの結果
type CreatePasswordResetValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreatePasswordResetValidator) Validate(ctx context.Context, input CreatePasswordResetValidatorInput) *CreatePasswordResetValidatorResult {
	formErrors := &session.FormErrors{}

	if strings.TrimSpace(input.Email) == "" {
		formErrors.AddFieldError("email", i18n.T(ctx, "password_reset_email_required"))
	}

	if formErrors.HasErrors() {
		return &CreatePasswordResetValidatorResult{FormErrors: formErrors}
	}

	return &CreatePasswordResetValidatorResult{}
}
