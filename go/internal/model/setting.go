package model

import "time"

// Setting はユーザー設定のドメインエンティティ
type Setting struct {
	ID                  SettingID
	UserID              UserID
	PrivacyPolicyAgreed bool
	HideRecordBody      bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
