package sign_up_username

import (
	"context"
	"regexp"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// usernameRegex ユーザー名の形式（1～20文字の半角英数字とアンダースコア）
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,20}$`)

// CreateValidator はユーザー登録のバリデーションを行う
type CreateValidator struct{}

// NewCreateValidator は CreateValidator を生成する
func NewCreateValidator() *CreateValidator {
	return &CreateValidator{}
}

// CreateValidatorInput はバリデーションの入力パラメータ
type CreateValidatorInput struct {
	Token    string
	Username string
}

// CreateValidatorResult はバリデーションの結果
type CreateValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateValidator) Validate(ctx context.Context, input CreateValidatorInput) *CreateValidatorResult {
	formErrors := &session.FormErrors{}

	// トークン必須チェック
	if input.Token == "" {
		formErrors.AddFieldError("token", i18n.T(ctx, "sign_up_username_error_token_missing"))
	}

	// ユーザー名必須チェック
	if input.Username == "" {
		formErrors.AddFieldError("username", i18n.T(ctx, "sign_up_username_error_username_required"))
	} else if !usernameRegex.MatchString(input.Username) {
		formErrors.AddFieldError("username", i18n.T(ctx, "sign_up_username_error_username_format"))
	}

	if formErrors.HasErrors() {
		return &CreateValidatorResult{FormErrors: formErrors}
	}

	return &CreateValidatorResult{}
}
