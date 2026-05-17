package validator

import (
	"context"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// SignUpCodeCreateValidator は新規登録確認コード検証のバリデーションを行う
type SignUpCodeCreateValidator struct{}

// NewSignUpCodeCreateValidator は SignUpCodeCreateValidator を生成する
func NewSignUpCodeCreateValidator() *SignUpCodeCreateValidator {
	return &SignUpCodeCreateValidator{}
}

// SignUpCodeCreateValidatorInput はバリデーションの入力パラメータ
type SignUpCodeCreateValidatorInput struct {
	Code string // 6桁の数字コード
}

// Validate はバリデーションを行う
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
