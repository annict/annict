package model

import (
	"encoding/json"
	"time"
)

// Session はセッションのドメインエンティティ
type Session struct {
	ID        int64
	SessionID string
	Data      json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SessionMaxAge is how long a session stays valid without being accessed. The
// Max-Age of the session cookie and the cutoff of the expired-session cleanup are
// both derived from it, so a session record is never deleted while the cookie that
// points at it is still valid.
//
// [Ja] SessionMaxAge はセッションがアクセスされないまま有効であり続ける期間。セッション
// Cookie の Max-Age と期限切れセッションのクリーンアップのカットオフをともにこの値から
// 導くため、Cookie がまだ有効なうちにレコードだけが消えることは起きない。
const SessionMaxAge = 30 * 24 * time.Hour
