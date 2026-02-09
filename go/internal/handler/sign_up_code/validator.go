package sign_up_code

import (
	"context"
	"regexp"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

var codeRegex = regexp.MustCompile(`^\d{6}$`)

// CreateValidator は新規登録確認コード検証のバリデーションを行う
type CreateValidator struct{}

// NewCreateValidator は CreateValidator を生成する
func NewCreateValidator() *CreateValidator {
	return &CreateValidator{}
}

// CreateValidatorInput はバリデーションの入力パラメータ
type CreateValidatorInput struct {
	Code string // 6桁の数字コード
}

// CreateValidatorResult はバリデーションの結果
type CreateValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateValidator) Validate(ctx context.Context, input CreateValidatorInput) *CreateValidatorResult {
	formErrors := &session.FormErrors{}

	// コードが空の場合
	if input.Code == "" {
		formErrors.AddFieldError("code", i18n.T(ctx, "sign_up_code_error_code_required"))
		return &CreateValidatorResult{FormErrors: formErrors}
	}

	// コードが6桁の数字でない場合
	if !codeRegex.MatchString(input.Code) {
		formErrors.AddFieldError("code", i18n.T(ctx, "sign_up_code_error_code_invalid_format"))
		return &CreateValidatorResult{FormErrors: formErrors}
	}

	return &CreateValidatorResult{}
}
