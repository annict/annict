package model

import (
	"database/sql"
	"time"
)

// UserTwoFactorAuthはユーザーが認証アプリに登録したTOTPの設定。
//
// SecretCiphertextは暗号化したTOTPの秘密鍵で、平文へ戻すのはauth.TwoFactorKeyだけが行う。
// LastUsedStepは最後に受け付けたコードのタイムステップで、0は一度も使っていないことを表す。
// EnabledAtは二要素認証を有効にした日時で、無効値は認証アプリへの登録の途中を表す。
type UserTwoFactorAuth struct {
	ID               UserTwoFactorAuthID
	UserID           UserID
	SecretCiphertext []byte
	LastUsedStep     int64
	EnabledAt        sql.NullTime
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsEnabledは二要素認証を有効にしているかを返す。
// 登録の途中の設定では、ログインでコードを求めない。
func (a *UserTwoFactorAuth) IsEnabled() bool {
	return a.EnabledAt.Valid
}
