package templates

import (
	"context"
	"net/http"
)

// DBSidebarOpenCookieNameはデスクトップ版Annict DBサイドバーの開閉設定を保存する
// Cookie名。
const DBSidebarOpenCookieName = "annict_db_sidebar_open"

// DBSidebarStateMiddlewareはデスクトップのサイドバー設定をリクエストコンテキストへ
// 保存し、クライアントJavaScriptの実行前からSSRされたサイドバーとトグルを正しい初期状態に
// する。Cookieが無い場合と値が不正な場合は開状態を既定とする。
func DBSidebarStateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		open := true
		if cookie, err := r.Cookie(DBSidebarOpenCookieName); err == nil && cookie.Value == "false" {
			open = false
		}
		next.ServeHTTP(w, r.WithContext(SetDBSidebarOpen(r.Context(), open)))
	})
}

// dbSidebarOpenContextKeyはデスクトップのサイドバー設定を保存する非公開context key。
type dbSidebarOpenContextKey struct{}

// SetDBSidebarOpenはデスクトップのサイドバー設定をコンテキストへ保存する。
func SetDBSidebarOpen(ctx context.Context, open bool) context.Context {
	return context.WithValue(ctx, dbSidebarOpenContextKey{}, open)
}

// IsDBSidebarOpenはデスクトップのサイドバー設定を返す。middlewareがcontextを設定して
// いない場合は開状態を返し、リクエストを介さない既存テストの描画挙動を維持する。
func IsDBSidebarOpen(ctx context.Context) bool {
	if open, ok := ctx.Value(dbSidebarOpenContextKey{}).(bool); ok {
		return open
	}
	return true
}
