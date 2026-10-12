package model

import "database/sql"

// Userは認証済みユーザーのドメインエンティティ
// Rails版のUser#role enumと対応: user: 0, admin: 1, editor: 2
type User struct {
	ID                  UserID
	Username            string
	Email               string
	Role                int32
	EncryptedPassword   string
	Locale              string
	TimeZone            string
	StripeSubscriberID  *StripeSubscriberID
	GumroadSubscriberID *GumroadSubscriberID
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

// IsAdminはユーザーが管理者かどうかを判定する
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsEditorはユーザーが編集者かどうかを判定する
func (u *User) IsEditor() bool {
	return u.Role == RoleEditor
}

// IsCommitterはユーザーが管理者または編集者かどうかを判定する
func (u *User) IsCommitter() bool {
	return u.IsAdmin() || u.IsEditor()
}

// UserDeletionTargetは退会の削除処理が対象のユーザーについて参照する属性。
// 削除処理は論理削除済みのユーザーだけを扱うため、論理削除の有無を持つ。
type UserDeletionTarget struct {
	ID                 UserID
	Deleted            bool
	StripeSubscriberID *StripeSubscriberID
}
