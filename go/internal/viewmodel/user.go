package viewmodel

import (
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/query"
)

// SidebarAvatarWidth はサイドバー用のアバター画像の幅（ピクセル）
const SidebarAvatarWidth = 40

// User はテンプレート表示用のユーザーデータです
type User struct {
	ID        int64
	Username  string
	AvatarURL string // 事前計算済みアバター画像URL（imgproxy経由）
}

// NewUserForSidebar はサイドバー表示用の viewmodel.User を作成します
// サイドバー用のアバター画像URL（40px, webp）を事前計算します
func NewUserForSidebar(row *query.GetUserByIDRow, helper *image.Helper) *User {
	if row == nil {
		return nil
	}

	var avatarURL string
	if row.ProfileImageData.Valid && helper != nil {
		avatarURL = helper.GetAvatarImageURL(row.ProfileImageData.String, SidebarAvatarWidth, "webp")
	}

	return &User{
		ID:        row.ID,
		Username:  row.Username,
		AvatarURL: avatarURL,
	}
}
