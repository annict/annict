// Package settings_two_factor_authは、二要素認証の設定・認証アプリの登録・リカバリーコードの表示の画面のテンプレートを定義する。
package settings_two_factor_auth

const (
	// ShowPathは二要素認証の設定画面のパス。有効にするフォームの送信先も兼ねる
	ShowPath = "/settings/two_factor_auth"
	// NewPathは認証アプリの登録画面のパス
	NewPath = "/settings/two_factor_auth/new"
	// SettingsBackPathは設定画面から戻る先のRails版の設定画面 (アカウント) のパス。
	// 二要素認証はアカウントの保護に関わる設定のため、同じ種類の設定が並ぶ画面へ戻す
	SettingsBackPath = "/settings/account"
)
