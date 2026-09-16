package validator

import (
	"context"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// SignInCreateValidatorはサインインフォームのバリデーションを行う
type SignInCreateValidator struct{}

// NewSignInCreateValidatorはSignInCreateValidatorを生成する
func NewSignInCreateValidator() *SignInCreateValidator {
	return &SignInCreateValidator{}
}

// SignInCreateValidatorInputはバリデーションの入力パラメータ
type SignInCreateValidatorInput struct {
	Email string
}

// Validateはフォームの形式バリデーションを行う
func (v *SignInCreateValidator) Validate(ctx context.Context, input SignInCreateValidatorInput) error {
	ve := model.NewValidationError()

	if strings.TrimSpace(input.Email) == "" {
		ve.AddField("email", i18n.T(ctx, "sign_in_error_email_required"))
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}
