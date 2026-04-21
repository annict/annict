package validator

import (
	"context"
	"regexp"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

var codeRegex = regexp.MustCompile(`^\d{6}$`)

// SignInCodeCreateValidator は6桁コード検証のバリデーションを行う
type SignInCodeCreateValidator struct{}

// NewSignInCodeCreateValidator は SignInCodeCreateValidator を生成する
func NewSignInCodeCreateValidator() *SignInCodeCreateValidator {
	return &SignInCodeCreateValidator{}
}

// SignInCodeCreateValidatorInput はバリデーションの入力パラメータ
type SignInCodeCreateValidatorInput struct {
	Code string // 6桁の数字コード
}

// Validate はバリデーションを行う
func (v *SignInCodeCreateValidator) Validate(ctx context.Context, input SignInCodeCreateValidatorInput) error {
	ve := model.NewValidationError()

	if input.Code == "" {
		ve.AddField("code", i18n.T(ctx, "sign_in_code_error_code_required"))
		return ve
	}

	if !codeRegex.MatchString(input.Code) {
		ve.AddField("code", i18n.T(ctx, "sign_in_code_error_code_invalid_format"))
		return ve
	}

	return nil
}
