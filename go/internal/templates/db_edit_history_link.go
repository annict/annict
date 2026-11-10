package templates

import "context"

// dbEditHistoryLinkContextKeyはAnnict DBのサイドバーに編集履歴の項目を出すかを保存する非公開context key。
type dbEditHistoryLinkContextKey struct{}

// SetDBEditHistoryLinkVisibleはAnnict DBのサイドバーに編集履歴の項目を出すかをコンテキストへ保存する。
func SetDBEditHistoryLinkVisible(ctx context.Context, visible bool) context.Context {
	return context.WithValue(ctx, dbEditHistoryLinkContextKey{}, visible)
}

// IsDBEditHistoryLinkVisibleはAnnict DBのサイドバーに編集履歴の項目を出すかを返す。middlewareがcontextを
// 設定していない場合は出さない。
func IsDBEditHistoryLinkVisible(ctx context.Context) bool {
	visible, _ := ctx.Value(dbEditHistoryLinkContextKey{}).(bool)
	return visible
}
