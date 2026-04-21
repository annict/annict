package validator

import (
	"context"
	"regexp"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// usernameRegex ユーザー名の形式（1～20文字の半角英数字とアンダースコア）
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,20}$`)

// SignUpUsernameCreateValidator はユーザー登録（ユーザー名設定）のバリデーションを行う
type SignUpUsernameCreateValidator struct{}

// NewSignUpUsernameCreateValidator は SignUpUsernameCreateValidator を生成する
func NewSignUpUsernameCreateValidator() *SignUpUsernameCreateValidator {
	return &SignUpUsernameCreateValidator{}
}

// SignUpUsernameCreateValidatorInput はバリデーションの入力パラメータ
type SignUpUsernameCreateValidatorInput struct {
	Token    string
	Username string
}

// Validate はバリデーションを行う
func (v *SignUpUsernameCreateValidator) Validate(ctx context.Context, input SignUpUsernameCreateValidatorInput) error {
	ve := model.NewValidationError()

	if input.Token == "" {
		ve.AddField("token", i18n.T(ctx, "sign_up_username_error_token_missing"))
	}

	if input.Username == "" {
		ve.AddField("username", i18n.T(ctx, "sign_up_username_error_username_required"))
	} else if !usernameRegex.MatchString(input.Username) {
		ve.AddField("username", i18n.T(ctx, "sign_up_username_error_username_format"))
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}
