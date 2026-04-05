package validator

import (
	"context"
	"regexp"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/session"
)

// usernameRegex ユーザー名の形式（1～20文字の半角英数字とアンダースコア）
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,20}$`)

// CreateSignUpUsernameValidator はユーザー登録（ユーザー名設定）のバリデーションを行う
type CreateSignUpUsernameValidator struct{}

// NewCreateSignUpUsernameValidator は CreateSignUpUsernameValidator を生成する
func NewCreateSignUpUsernameValidator() *CreateSignUpUsernameValidator {
	return &CreateSignUpUsernameValidator{}
}

// CreateSignUpUsernameValidatorInput はバリデーションの入力パラメータ
type CreateSignUpUsernameValidatorInput struct {
	Token    string
	Username string
}

// CreateSignUpUsernameValidatorResult はバリデーションの結果
type CreateSignUpUsernameValidatorResult struct {
	FormErrors *session.FormErrors
}

// Validate はバリデーションを行う
func (v *CreateSignUpUsernameValidator) Validate(ctx context.Context, input CreateSignUpUsernameValidatorInput) *CreateSignUpUsernameValidatorResult {
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
		return &CreateSignUpUsernameValidatorResult{FormErrors: formErrors}
	}

	return &CreateSignUpUsernameValidatorResult{}
}
