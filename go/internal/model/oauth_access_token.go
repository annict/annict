package model

import "time"

// OAuthAccessTokenはアクセストークンを表すドメインモデル。
// OAuthで発行したトークンと、設定ページで発行した個人用アクセストークンの両方を表す
type OAuthAccessToken struct {
	ID              OAuthAccessTokenID
	ResourceOwnerID UserID
	// Personalは個人用アクセストークンかどうか。OAuthのアプリケーションに結び付かないトークンが当たる
	Personal bool
	// ExpiresInは発行から失効までの長さ。nilなら期限が無い
	ExpiresIn *time.Duration
	RevokedAt *time.Time
	CreatedAt time.Time
	// Scopesはトークンに付けたスコープ (`read`・`write`)
	Scopes []string
}

// IsUsableAtは、nowの時点でトークンを使えるか (失効しておらず、期限が切れていないか) を返す
func (t *OAuthAccessToken) IsUsableAt(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	if t.ExpiresIn != nil && !now.Before(t.CreatedAt.Add(*t.ExpiresIn)) {
		return false
	}
	return true
}
