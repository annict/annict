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

// AnonymousSessionMaxAgeはログインしていないセッションの行を残す期間。未ログインの
// セッションの大半はCSRFトークンのためだけに作られて二度と使われないため、ログイン済みの
// セッション (SessionMaxAge) より短い期間で削除する。CookieのMax-Ageはこの値から導かず
// SessionMaxAgeのままにする。行が消えたあとにCookieが残っていても、新しいセッションとして
// 扱われるだけでエラーにはならないため。
const AnonymousSessionMaxAge = 24 * time.Hour
