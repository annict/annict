package password

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// UpdateValidator はパスワード更新フォームのバリデーションを行う
type UpdateValidator struct{}

// NewUpdateValidator は UpdateValidator を生成する
func NewUpdateValidator() *UpdateValidator {
	return &UpdateValidator{}
}

// UpdateValidatorInput はバリデーションの入力パラメータ
type UpdateValidatorInput struct {
	Token                string
	Password             string
	PasswordConfirmation string
}

// UpdateValidatorResult はバリデーションの結果
type UpdateValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *UpdateValidator) Validate(ctx context.Context, input UpdateValidatorInput) *UpdateValidatorResult {
	formErrors := &session.FormErrors{}

	if strings.TrimSpace(input.Token) == "" {
		formErrors.AddFieldError("token", i18n.T(ctx, "password_reset_token_invalid"))
	}

	if strings.TrimSpace(input.Password) == "" {
		formErrors.AddFieldError("password", i18n.T(ctx, "password_reset_password_required"))
	}

	if strings.TrimSpace(input.PasswordConfirmation) == "" {
		formErrors.AddFieldError("password_confirmation", i18n.T(ctx, "password_reset_password_confirmation_required"))
	}

	if strings.TrimSpace(input.Password) != "" && strings.TrimSpace(input.PasswordConfirmation) != "" && input.Password != input.PasswordConfirmation {
		formErrors.AddFieldError("password_confirmation", i18n.T(ctx, "password_reset_password_mismatch"))
	}

	if formErrors.HasErrors() {
		return &UpdateValidatorResult{FormErrors: formErrors}
	}

	return &UpdateValidatorResult{}
}
