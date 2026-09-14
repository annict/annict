package validator

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// PasswordResetCreateValidatorはパスワードリセット申請フォームのバリデーションを行う
type PasswordResetCreateValidator struct{}

// NewPasswordResetCreateValidatorはPasswordResetCreateValidatorを生成する
func NewPasswordResetCreateValidator() *PasswordResetCreateValidator {
	return &PasswordResetCreateValidator{}
}

// PasswordResetCreateValidatorInputはバリデーションの入力パラメータ
type PasswordResetCreateValidatorInput struct {
	Email string
}

// Validateはバリデーションを行う
func (v *PasswordResetCreateValidator) Validate(ctx context.Context, input PasswordResetCreateValidatorInput) error {
	ve := model.NewValidationError()

	if strings.TrimSpace(input.Email) == "" {
		ve.AddField("email", i18n.T(ctx, "password_reset_email_required"))
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}
