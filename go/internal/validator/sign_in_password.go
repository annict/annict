package validator

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreateSignInPasswordValidator はパスワードログインのバリデーションを行う
type CreateSignInPasswordValidator struct{}

// NewCreateSignInPasswordValidator は CreateSignInPasswordValidator を生成する
func NewCreateSignInPasswordValidator() *CreateSignInPasswordValidator {
	return &CreateSignInPasswordValidator{}
}

// CreateSignInPasswordValidatorInput はバリデーションの入力パラメータ
type CreateSignInPasswordValidatorInput struct {
	Password string
}

// CreateSignInPasswordValidatorResult はバリデーションの結果
type CreateSignInPasswordValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateSignInPasswordValidator) Validate(ctx context.Context, input CreateSignInPasswordValidatorInput) *CreateSignInPasswordValidatorResult {
	formErrors := &session.FormErrors{}

	if strings.TrimSpace(input.Password) == "" {
		formErrors.AddFieldError("password", i18n.T(ctx, "sign_in_error_password_required"))
		return &CreateSignInPasswordValidatorResult{FormErrors: formErrors}
	}

	return &CreateSignInPasswordValidatorResult{}
}
