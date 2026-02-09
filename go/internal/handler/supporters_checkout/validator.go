package supporters_checkout

import (
	"context"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreateValidator はCheckoutセッション作成のバリデーションを行う
type CreateValidator struct{}

// NewCreateValidator は CreateValidator を生成する
func NewCreateValidator() *CreateValidator {
	return &CreateValidator{}
}

// CreateValidatorInput はバリデーションの入力パラメータ
type CreateValidatorInput struct {
	Plan string
}

// CreateValidatorResult はバリデーションの結果
type CreateValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateValidator) Validate(ctx context.Context, input CreateValidatorInput) *CreateValidatorResult {
	formErrors := &session.FormErrors{}

	if input.Plan != "monthly" && input.Plan != "yearly" {
		formErrors.AddFieldError("plan", i18n.T(ctx, "supporters_checkout_invalid_plan"))
	}

	if formErrors.HasErrors() {
		return &CreateValidatorResult{FormErrors: formErrors}
	}

	return &CreateValidatorResult{}
}
