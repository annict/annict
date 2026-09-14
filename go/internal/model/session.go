package model

import (
	"encoding/json"
	"time"
)

// Sessionはセッションのドメインエンティティ
type Session struct {
	ID        int64
	SessionID string
	Data      json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SessionMaxAgeはセッションがアクセスされないまま有効であり続ける期間。セッション
// CookieのMax-Ageと期限切れセッションのクリーンアップのカットオフをともにこの値から
// 導くため、Cookieがまだ有効なうちにレコードだけが消えることは起きない。
const SessionMaxAge = 30 * 24 * time.Hour
