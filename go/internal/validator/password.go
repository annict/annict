package validator

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// PasswordUpdateValidator はパスワード更新フォームのバリデーションを行う
type PasswordUpdateValidator struct{}

// NewPasswordUpdateValidator は PasswordUpdateValidator を生成する
func NewPasswordUpdateValidator() *PasswordUpdateValidator {
	return &PasswordUpdateValidator{}
}

// PasswordUpdateValidatorInput はバリデーションの入力パラメータ
type PasswordUpdateValidatorInput struct {
	Token                string
	Password             string
	PasswordConfirmation string
}

// Validate はバリデーションを行う
func (v *PasswordUpdateValidator) Validate(ctx context.Context, input PasswordUpdateValidatorInput) error {
	ve := model.NewValidationError()

	if strings.TrimSpace(input.Token) == "" {
		ve.AddField("token", i18n.T(ctx, "password_reset_token_invalid"))
	}

	if strings.TrimSpace(input.Password) == "" {
		ve.AddField("password", i18n.T(ctx, "password_reset_password_required"))
	}

	if strings.TrimSpace(input.PasswordConfirmation) == "" {
		ve.AddField("password_confirmation", i18n.T(ctx, "password_reset_password_confirmation_required"))
	}

	if strings.TrimSpace(input.Password) != "" && strings.TrimSpace(input.PasswordConfirmation) != "" && input.Password != input.PasswordConfirmation {
		ve.AddField("password_confirmation", i18n.T(ctx, "password_reset_password_mismatch"))
	}

	if ve.HasErrors() {
		return ve
	}

	// パスワード強度チェック
	if err := auth.ValidatePasswordStrength(ctx, input.Password); err != nil {
		ve.AddField("password", err.Error())
		return ve
	}

	return nil
}
