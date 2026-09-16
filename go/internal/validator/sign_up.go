package validator

import (
	"context"
	"net/mail"
	"strings"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// SignUpCreateValidatorは新規登録フォームのバリデーションを行う
type SignUpCreateValidator struct{}

// NewSignUpCreateValidatorはSignUpCreateValidatorを生成する
func NewSignUpCreateValidator() *SignUpCreateValidator {
	return &SignUpCreateValidator{}
}

// SignUpCreateValidatorInputはバリデーションの入力パラメータ
type SignUpCreateValidatorInput struct {
	Email string
}

// Validateはフォームの形式バリデーションを行う
func (v *SignUpCreateValidator) Validate(ctx context.Context, input SignUpCreateValidatorInput) error {
	ve := model.NewValidationError()

	if strings.TrimSpace(input.Email) == "" {
		ve.AddField("email", i18n.T(ctx, "sign_up_error_email_required"))
		return ve
	}

	if _, err := mail.ParseAddress(input.Email); err != nil {
		ve.AddField("email", i18n.T(ctx, "sign_up_error_email_invalid"))
		return ve
	}

	return nil
}
