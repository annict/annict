package validator

import (
	"context"
	"regexp"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

var codeRegex = regexp.MustCompile(`^\d{6}$`)

// CreateSignInCodeValidator は6桁コード検証のバリデーションを行う
type CreateSignInCodeValidator struct{}

// NewCreateSignInCodeValidator は CreateSignInCodeValidator を生成する
func NewCreateSignInCodeValidator() *CreateSignInCodeValidator {
	return &CreateSignInCodeValidator{}
}

// CreateSignInCodeValidatorInput はバリデーションの入力パラメータ
type CreateSignInCodeValidatorInput struct {
	Code string // 6桁の数字コード
}

// CreateSignInCodeValidatorResult はバリデーションの結果
type CreateSignInCodeValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateSignInCodeValidator) Validate(ctx context.Context, input CreateSignInCodeValidatorInput) *CreateSignInCodeValidatorResult {
	formErrors := &session.FormErrors{}

	// コードが空の場合
	if input.Code == "" {
		formErrors.AddFieldError("code", i18n.T(ctx, "sign_in_code_error_code_required"))
		return &CreateSignInCodeValidatorResult{FormErrors: formErrors}
	}

	// コードが6桁の数字でない場合
	if !codeRegex.MatchString(input.Code) {
		formErrors.AddFieldError("code", i18n.T(ctx, "sign_in_code_error_code_invalid_format"))
		return &CreateSignInCodeValidatorResult{FormErrors: formErrors}
	}

	return &CreateSignInCodeValidatorResult{}
}
