package validator

import (
	"context"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// SupportersCheckoutCreateValidator はCheckoutセッション作成のバリデーションを行う
type SupportersCheckoutCreateValidator struct{}

// NewSupportersCheckoutCreateValidator は SupportersCheckoutCreateValidator を生成する
func NewSupportersCheckoutCreateValidator() *SupportersCheckoutCreateValidator {
	return &SupportersCheckoutCreateValidator{}
}

// SupportersCheckoutCreateValidatorInput はバリデーションの入力パラメータ
type SupportersCheckoutCreateValidatorInput struct {
	Plan string
}

// Validate はバリデーションを行う
func (v *SupportersCheckoutCreateValidator) Validate(ctx context.Context, input SupportersCheckoutCreateValidatorInput) error {
	ve := model.NewValidationError()

	if input.Plan != "monthly" && input.Plan != "yearly" {
		ve.AddField("plan", i18n.T(ctx, "supporters_checkout_invalid_plan"))
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}
