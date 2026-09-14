package model

import "time"

// Settingはユーザー設定のドメインエンティティ
type Setting struct {
	ID                  SettingID
	UserID              UserID
	PrivacyPolicyAgreed bool
	HideRecordBody      bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
