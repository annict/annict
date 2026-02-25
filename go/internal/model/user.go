package model

import "database/sql"

// User は認証済みユーザーのドメインエンティティ
// Rails版の User#role enum と対応: user: 0, admin: 1, editor: 2
type User struct {
	ID                  int64
	Username            string
	Email               string
	Role                int32
	EncryptedPassword   string
	Locale              string
	TimeZone            string
	StripeSubscriberID  sql.NullInt64
	GumroadSubscriberID sql.NullInt64
	NotificationsCount  int32
	CreatedAt           sql.NullTime
	UpdatedAt           sql.NullTime
	ProfileImageData    sql.NullString
}

// ロール定数
const (
	RoleUser   int32 = 0
	RoleAdmin  int32 = 1
	RoleEditor int32 = 2
)

// IsAdmin はユーザーが管理者かどうかを判定する
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsEditor はユーザーが編集者かどうかを判定する
func (u *User) IsEditor() bool {
	return u.Role == RoleEditor
}

// IsCommitter はユーザーが管理者または編集者かどうかを判定する
func (u *User) IsCommitter() bool {
	return u.IsAdmin() || u.IsEditor()
}
