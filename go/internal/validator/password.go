package validator

import (
	"context"
	"errors"
	"log/slog"
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

	// パスワード強度チェック（sentinel error から i18n 翻訳を解決）
	if err := auth.ValidatePasswordStrength(input.Password); err != nil {
		switch {
		case errors.Is(err, auth.ErrPasswordTooShort):
			ve.AddField("password", i18n.T(ctx, "password_strength_min_length", map[string]any{
				"MinLength": auth.MinPasswordLength,
			}))
		case errors.Is(err, auth.ErrPasswordTooLong):
			ve.AddField("password", i18n.T(ctx, "password_strength_max_length", map[string]any{
				"MaxLength": auth.MaxPasswordLength,
			}))
		case errors.Is(err, auth.ErrPasswordInvalidChars):
			ve.AddField("password", i18n.T(ctx, "password_strength_invalid_chars"))
		default:
			slog.ErrorContext(ctx, "auth.ValidatePasswordStrength から未知の sentinel error が返りました。validator 側に switch case の追加が必要です", "error", err)
			return err
		}
		return ve
	}

	return nil
}
