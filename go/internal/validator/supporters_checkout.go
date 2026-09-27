package validator

import (
	"context"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// SupportersCheckoutCreateValidatorはCheckoutセッション作成のバリデーションを行う
type SupportersCheckoutCreateValidator struct{}

// NewSupportersCheckoutCreateValidatorはSupportersCheckoutCreateValidatorを生成する
func NewSupportersCheckoutCreateValidator() *SupportersCheckoutCreateValidator {
	return &SupportersCheckoutCreateValidator{}
}

// SupportersCheckoutCreateValidatorInputはバリデーションの入力パラメータ
type SupportersCheckoutCreateValidatorInput struct {
	Plan string
}

// Validateはバリデーションを行う
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
