// Package two_factor_auth_notificationは、二要素認証の設定の変更を本人に知らせるメールのテンプレートを定義する。
package two_factor_auth_notification

// Kindはメールで知らせる変更の種類。テンプレートはModelに依存できないため、
// email.TwoFactorAuthNotificationSenderがmodel.TwoFactorAuthNotificationEventから対応付ける。
type Kind int

const (
	// KindEnabledは本人が二要素認証を有効にしたことを表す
	KindEnabled Kind = iota
	// KindDisabledは本人が二要素認証を無効にしたことを表す
	KindDisabled
	// KindDisabledByAdminは管理者が二要素認証を無効にしたことを表す
	KindDisabledByAdmin
)
