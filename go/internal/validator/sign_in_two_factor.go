package validator

import (
	"context"

	"github.com/annict/annict/go/internal/model"
)

// SignInTwoFactorCreateValidatorは、ログインで認証アプリのコードを入力するフォームのコードを検証する。
//
// コードの照合は行わない。秘密鍵の復号と、使ったタイムステップの記録が要るため、UseCaseが行う。
type SignInTwoFactorCreateValidator struct{}

// NewSignInTwoFactorCreateValidatorはSignInTwoFactorCreateValidatorを生成する
func NewSignInTwoFactorCreateValidator() *SignInTwoFactorCreateValidator {
	return &SignInTwoFactorCreateValidator{}
}

// SignInTwoFactorCreateValidatorInputはバリデーションの入力パラメータ
type SignInTwoFactorCreateValidatorInput struct {
	// Codeはauth.NormalizeTOTPCodeで空白とハイフンを取り除いたコード
	Code string
}

// Validateはコードが入力され、半角数字6桁であることを確かめる。失敗時は *model.ValidationErrorを返す
func (v *SignInTwoFactorCreateValidator) Validate(ctx context.Context, input SignInTwoFactorCreateValidatorInput) error {
	ve := model.NewValidationError()
	addTOTPCodeErrors(ctx, ve, input.Code)

	if ve.HasErrors() {
		return ve
	}
	return nil
}
