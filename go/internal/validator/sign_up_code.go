package validator

import (
	"context"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// CreateSignUpCodeValidator は新規登録確認コード検証のバリデーションを行う
type CreateSignUpCodeValidator struct{}

// NewCreateSignUpCodeValidator は CreateSignUpCodeValidator を生成する
func NewCreateSignUpCodeValidator() *CreateSignUpCodeValidator {
	return &CreateSignUpCodeValidator{}
}

// CreateSignUpCodeValidatorInput はバリデーションの入力パラメータ
type CreateSignUpCodeValidatorInput struct {
	Code string // 6桁の数字コード
}

// CreateSignUpCodeValidatorResult はバリデーションの結果
type CreateSignUpCodeValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateSignUpCodeValidator) Validate(ctx context.Context, input CreateSignUpCodeValidatorInput) *CreateSignUpCodeValidatorResult {
	formErrors := &session.FormErrors{}

	// コードが空の場合
	if input.Code == "" {
		formErrors.AddFieldError("code", i18n.T(ctx, "sign_up_code_error_code_required"))
		return &CreateSignUpCodeValidatorResult{FormErrors: formErrors}
	}

	// コードが6桁の数字でない場合
	if !codeRegex.MatchString(input.Code) {
		formErrors.AddFieldError("code", i18n.T(ctx, "sign_up_code_error_code_invalid_format"))
		return &CreateSignUpCodeValidatorResult{FormErrors: formErrors}
	}

	return &CreateSignUpCodeValidatorResult{}
}
