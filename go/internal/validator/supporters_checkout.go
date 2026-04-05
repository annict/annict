package validator

import (
	"context"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreateSupportersCheckoutValidator はCheckoutセッション作成のバリデーションを行う
type CreateSupportersCheckoutValidator struct{}

// NewCreateSupportersCheckoutValidator は CreateSupportersCheckoutValidator を生成する
func NewCreateSupportersCheckoutValidator() *CreateSupportersCheckoutValidator {
	return &CreateSupportersCheckoutValidator{}
}

// CreateSupportersCheckoutValidatorInput はバリデーションの入力パラメータ
type CreateSupportersCheckoutValidatorInput struct {
	Plan string
}

// CreateSupportersCheckoutValidatorResult はバリデーションの結果
type CreateSupportersCheckoutValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateSupportersCheckoutValidator) Validate(ctx context.Context, input CreateSupportersCheckoutValidatorInput) *CreateSupportersCheckoutValidatorResult {
	formErrors := &session.FormErrors{}

	if input.Plan != "monthly" && input.Plan != "yearly" {
		formErrors.AddFieldError("plan", i18n.T(ctx, "supporters_checkout_invalid_plan"))
	}

	if formErrors.HasErrors() {
		return &CreateSupportersCheckoutValidatorResult{FormErrors: formErrors}
	}

	return &CreateSupportersCheckoutValidatorResult{}
}
