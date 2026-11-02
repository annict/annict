package validator

import (
	"context"

	"github.com/annict/annict/go/internal/auth"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/model"
)

// SettingsTwoFactorAuthCreateValidatorは、二要素認証を有効にするフォームのコードを検証する。
//
// コードの照合は行わない。秘密鍵の復号と、使ったタイムステップの記録が要るため、UseCaseが行う。
type SettingsTwoFactorAuthCreateValidator struct{}

// NewSettingsTwoFactorAuthCreateValidatorはSettingsTwoFactorAuthCreateValidatorを生成する
func NewSettingsTwoFactorAuthCreateValidator() *SettingsTwoFactorAuthCreateValidator {
	return &SettingsTwoFactorAuthCreateValidator{}
}

// SettingsTwoFactorAuthCreateValidatorInputはバリデーションの入力パラメータ
type SettingsTwoFactorAuthCreateValidatorInput struct {
	// Codeはauth.NormalizeTOTPCodeで空白とハイフンを取り除いたコード
	Code string
}

// Validateはコードが入力され、半角数字6桁であることを確かめる。失敗時は *model.ValidationErrorを返す
func (v *SettingsTwoFactorAuthCreateValidator) Validate(ctx context.Context, input SettingsTwoFactorAuthCreateValidatorInput) error {
	ve := model.NewValidationError()
	addTOTPCodeErrors(ctx, ve, input.Code)

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// addTOTPCodeErrorsは、認証アプリのコードが入力され、半角数字6桁であることを確かめ、誤りをコードの欄に加える。
// 有効にするときとログインするときで、同じ形のコードを受け付けるため共有する。
func addTOTPCodeErrors(ctx context.Context, ve *model.ValidationError, code string) {
	switch {
	case code == "":
		ve.AddField("code", i18n.T(ctx, "validation_required"))
	case !auth.IsTOTPCodeFormat(code):
		ve.AddField("code", i18n.T(ctx, "validation_totp_code_format"))
	}
}

// SettingsTwoFactorAuthDeleteValidatorは、二要素認証を無効にするフォームの本人確認の入力を検証する。
//
// 入力はパスワード・認証アプリのコード・リカバリーコードのいずれかで、照合は行わない。
// 照合にはTOTPの秘密鍵の復号・リカバリーコードの消費・パスワードのハッシュが要るため、UseCaseが行う。
type SettingsTwoFactorAuthDeleteValidator struct{}

// NewSettingsTwoFactorAuthDeleteValidatorはSettingsTwoFactorAuthDeleteValidatorを生成する
func NewSettingsTwoFactorAuthDeleteValidator() *SettingsTwoFactorAuthDeleteValidator {
	return &SettingsTwoFactorAuthDeleteValidator{}
}

// SettingsTwoFactorAuthDeleteValidatorInputはバリデーションの入力パラメータ
type SettingsTwoFactorAuthDeleteValidatorInput struct {
	// Credentialは入力されたままのパスワード・認証アプリのコード・リカバリーコードのいずれか
	Credential string
}

// Validateは本人確認の入力があることを確かめる。失敗時は *model.ValidationErrorを返す
//
// 形式は確かめない。どの形でもなければパスワードとして照合するため、パスワードの長さのポリシーも当てはめない。
func (v *SettingsTwoFactorAuthDeleteValidator) Validate(ctx context.Context, input SettingsTwoFactorAuthDeleteValidatorInput) error {
	if input.Credential == "" {
		ve := model.NewValidationError()
		ve.AddField("credential", i18n.T(ctx, "validation_required"))
		return ve
	}

	return nil
}
