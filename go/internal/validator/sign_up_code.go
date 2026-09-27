package validator

import (
	"context"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// SignUpCodeCreateValidatorは新規登録確認コード検証のバリデーションを行う
type SignUpCodeCreateValidator struct{}

// NewSignUpCodeCreateValidatorはSignUpCodeCreateValidatorを生成する
func NewSignUpCodeCreateValidator() *SignUpCodeCreateValidator {
	return &SignUpCodeCreateValidator{}
}

// SignUpCodeCreateValidatorInputはバリデーションの入力パラメータ
type SignUpCodeCreateValidatorInput struct {
	Code string // 6桁の数字コード
}

// Validateはバリデーションを行う
func (v *SignUpCodeCreateValidator) Validate(ctx context.Context, input SignUpCodeCreateValidatorInput) error {
	ve := model.NewValidationError()

	if input.Code == "" {
		ve.AddField("code", i18n.T(ctx, "sign_up_code_error_code_required"))
		return ve
	}

	if !codeRegex.MatchString(input.Code) {
		ve.AddField("code", i18n.T(ctx, "sign_up_code_error_code_invalid_format"))
		return ve
	}

	return nil
}
