package validator

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreateSignInValidator はサインインフォームのバリデーションを行う
type CreateSignInValidator struct{}

// NewCreateSignInValidator は CreateSignInValidator を生成する
func NewCreateSignInValidator() *CreateSignInValidator {
	return &CreateSignInValidator{}
}

// CreateSignInValidatorInput はバリデーションの入力パラメータ
type CreateSignInValidatorInput struct {
	Email string
}

// CreateSignInValidatorResult はバリデーションの結果
type CreateSignInValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はフォームの形式バリデーションを行う
func (v *CreateSignInValidator) Validate(ctx context.Context, input CreateSignInValidatorInput) *CreateSignInValidatorResult {
	formErrors := &session.FormErrors{}

	if strings.TrimSpace(input.Email) == "" {
		formErrors.AddFieldError("email", i18n.T(ctx, "sign_in_email_required"))
	}

	if formErrors.HasErrors() {
		return &CreateSignInValidatorResult{FormErrors: formErrors}
	}
	return &CreateSignInValidatorResult{}
}
