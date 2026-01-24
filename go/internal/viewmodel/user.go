package viewmodel

import (
	"github.com/annict/annict/go/internal/image"
	"github.com/annict/annict/go/internal/repository"
)

// サイドバー用アバター画像サイズ（Retina対応で表示サイズの2倍）
const sidebarAvatarImageSize = 100 // 50px × 2

// User はテンプレート表示用のユーザーデータです
type User struct {
	ID                 int64
	Username           string
	AvatarURL          string // サイドバー用アバター画像URL（50px表示、100px画像）
	NotificationsCount int32  // 未読通知数
}

// NewUserForSidebar はサイドバー表示用の viewmodel.User を作成します
func NewUserForSidebar(row *repository.User, helper *image.Helper) *User {
	if row == nil {
		return nil
	}

	var avatarURL string
	if row.ProfileImageData.Valid && helper != nil {
		avatarURL = helper.GetAvatarImageURL(row.ProfileImageData.String, sidebarAvatarImageSize, "webp")
	}

	return &User{
		ID:                 row.ID,
		Username:           row.Username,
		AvatarURL:          avatarURL,
		NotificationsCount: row.NotificationsCount,
	}
}
