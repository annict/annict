package validator

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// UpdatePasswordValidator はパスワード更新フォームのバリデーションを行う
type UpdatePasswordValidator struct{}

// NewUpdatePasswordValidator は UpdatePasswordValidator を生成する
func NewUpdatePasswordValidator() *UpdatePasswordValidator {
	return &UpdatePasswordValidator{}
}

// UpdatePasswordValidatorInput はバリデーションの入力パラメータ
type UpdatePasswordValidatorInput struct {
	Token                string
	Password             string
	PasswordConfirmation string
}

// UpdatePasswordValidatorResult はバリデーションの結果
type UpdatePasswordValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *UpdatePasswordValidator) Validate(ctx context.Context, input UpdatePasswordValidatorInput) *UpdatePasswordValidatorResult {
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
		return &UpdatePasswordValidatorResult{FormErrors: formErrors}
	}

	// パスワード強度チェック
	if err := auth.ValidatePasswordStrength(ctx, input.Password); err != nil {
		formErrors.AddFieldError("password", err.Error())
		return &UpdatePasswordValidatorResult{FormErrors: formErrors}
	}

	return &UpdatePasswordValidatorResult{}
}
