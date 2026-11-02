package model

// TwoFactorAuthNotificationEventは、二要素認証の設定の変更を本人に知らせるメールの種類。
// ジョブの引数に文字列のまま保存するため、値を変えるとキューに残ったジョブを送れなくなる。
type TwoFactorAuthNotificationEvent string

const (
	// TwoFactorAuthNotificationEventEnabledは本人が二要素認証を有効にしたことを表す
	TwoFactorAuthNotificationEventEnabled TwoFactorAuthNotificationEvent = "enabled"
	// TwoFactorAuthNotificationEventDisabledは本人が二要素認証を無効にしたことを表す
	TwoFactorAuthNotificationEventDisabled TwoFactorAuthNotificationEvent = "disabled"
	// TwoFactorAuthNotificationEventDisabledByAdminは管理者が二要素認証を無効にしたことを表す
	TwoFactorAuthNotificationEventDisabledByAdmin TwoFactorAuthNotificationEvent = "disabled_by_admin"
)

// IsValidは知っている種類かを返す。キューから取り出したジョブの引数を確かめるのに使う。
func (e TwoFactorAuthNotificationEvent) IsValid() bool {
	switch e {
	case TwoFactorAuthNotificationEventEnabled, TwoFactorAuthNotificationEventDisabled, TwoFactorAuthNotificationEventDisabledByAdmin:
		return true
	default:
		return false
	}
}
