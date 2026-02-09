package sign_in_password

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreateValidator はパスワードログインのバリデーションを行う
type CreateValidator struct{}

// NewCreateValidator は CreateValidator を生成する
func NewCreateValidator() *CreateValidator {
	return &CreateValidator{}
}

// CreateValidatorInput はバリデーションの入力パラメータ
type CreateValidatorInput struct {
	Password string
}

// CreateValidatorResult はバリデーションの結果
type CreateValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateValidator) Validate(ctx context.Context, input CreateValidatorInput) *CreateValidatorResult {
	formErrors := &session.FormErrors{}

	if strings.TrimSpace(input.Password) == "" {
		formErrors.AddFieldError("password", i18n.T(ctx, "sign_in_error_password_required"))
		return &CreateValidatorResult{FormErrors: formErrors}
	}

	return &CreateValidatorResult{}
}
