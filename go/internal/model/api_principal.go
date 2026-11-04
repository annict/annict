package model

import "slices"

// APITokenKindはREST API V2のリクエストに使ったトークンの種類
type APITokenKind int

const (
	// APITokenKindOAuthAccessTokenはOAuthで発行したアクセストークン
	APITokenKindOAuthAccessToken APITokenKind = iota + 1
	// APITokenKindPersonalAccessTokenは設定ページで発行した個人用アクセストークン
	APITokenKindPersonalAccessToken
)

// APIPrincipalはREST API V2のリクエストを、照合したトークンで送った呼び出し主体
type APIPrincipal struct {
	// UserIDはトークンの持ち主の利用者のID
	UserID    UserID
	TokenKind APITokenKind
	// Scopesはトークンに付けたスコープ
	Scopes []string
	// RESTAPIV2Enabledは、トークンの持ち主にフィーチャーフラグ `rest_api_v2` が有効かどうか
	RESTAPIV2Enabled bool
}

// HasScopeはトークンがscopeを持つかを返す
func (p *APIPrincipal) HasScope(scope string) bool {
	return slices.Contains(p.Scopes, scope)
}
