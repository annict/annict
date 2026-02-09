package password_reset

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreateValidator はパスワードリセット申請フォームのバリデーションを行う
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

	if strings.TrimSpace(input.Email) == "" {
		formErrors.AddFieldError("email", i18n.T(ctx, "password_reset_email_required"))
	}

	if formErrors.HasErrors() {
		return &CreateValidatorResult{FormErrors: formErrors}
	}

	return &CreateValidatorResult{}
}
