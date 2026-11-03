package validator

import (
	"context"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// SignInTwoFactorRecoveryCreateValidatorは、ログインでリカバリーコードを入力するフォームのコードを検証する。
//
// コードの照合は行わない。ダイジェストとの照合と消費は、同時に送られた要求と競うため、UseCaseがトランザクションの中で行う。
type SignInTwoFactorRecoveryCreateValidator struct{}

// NewSignInTwoFactorRecoveryCreateValidatorはSignInTwoFactorRecoveryCreateValidatorを生成する
func NewSignInTwoFactorRecoveryCreateValidator() *SignInTwoFactorRecoveryCreateValidator {
	return &SignInTwoFactorRecoveryCreateValidator{}
}

// SignInTwoFactorRecoveryCreateValidatorInputはバリデーションの入力パラメータ
type SignInTwoFactorRecoveryCreateValidatorInput struct {
	// Codeはauth.NormalizeRecoveryCodeで区切りを取り除いて小文字にしたコード
	Code string
}

// Validateはコードが入力され、発行するリカバリーコードの形であることを確かめる。失敗時は *model.ValidationErrorを返す
func (v *SignInTwoFactorRecoveryCreateValidator) Validate(ctx context.Context, input SignInTwoFactorRecoveryCreateValidatorInput) error {
	ve := model.NewValidationError()

	switch {
	case input.Code == "":
		ve.AddField("code", i18n.T(ctx, "validation_required"))
	case !auth.IsRecoveryCodeFormat(input.Code):
		ve.AddField("code", i18n.T(ctx, "validation_recovery_code_format"))
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}
