package templates

import (
	"context"
	"net/http"
)

// DBSidebarOpenCookieName is the cookie that stores the desktop Annict DB sidebar preference.
//
// [Ja] DBSidebarOpenCookieName はデスクトップ版 Annict DB サイドバーの開閉設定を保存する
// Cookie 名。
const DBSidebarOpenCookieName = "annict_db_sidebar_open"

// DBSidebarStateMiddleware stores the desktop sidebar preference in the request context so the
// server-rendered sidebar and toggle have the correct initial state before client JavaScript runs.
// Missing and invalid cookie values default to open.
//
// [Ja] DBSidebarStateMiddleware はデスクトップのサイドバー設定をリクエストコンテキストへ
// 保存し、クライアント JavaScript の実行前から SSR されたサイドバーとトグルを正しい初期状態に
// する。Cookie が無い場合と値が不正な場合は開状態を既定とする。
func DBSidebarStateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		open := true
		if cookie, err := r.Cookie(DBSidebarOpenCookieName); err == nil && cookie.Value == "false" {
			open = false
		}
		next.ServeHTTP(w, r.WithContext(SetDBSidebarOpen(r.Context(), open)))
	})
}

// dbSidebarOpenContextKey is the private context key for the desktop sidebar preference.
//
// [Ja] dbSidebarOpenContextKey はデスクトップのサイドバー設定を保存する非公開 context key。
type dbSidebarOpenContextKey struct{}

// SetDBSidebarOpen stores the desktop sidebar preference in the context.
//
// [Ja] SetDBSidebarOpen はデスクトップのサイドバー設定をコンテキストへ保存する。
func SetDBSidebarOpen(ctx context.Context, open bool) context.Context {
	return context.WithValue(ctx, dbSidebarOpenContextKey{}, open)
}

// IsDBSidebarOpen returns the desktop sidebar preference. It defaults to open when middleware has
// not populated the context, preserving the existing rendering behavior in non-request tests.
//
// [Ja] IsDBSidebarOpen はデスクトップのサイドバー設定を返す。middleware が context を設定して
// いない場合は開状態を返し、リクエストを介さない既存テストの描画挙動を維持する。
func IsDBSidebarOpen(ctx context.Context) bool {
	if open, ok := ctx.Value(dbSidebarOpenContextKey{}).(bool); ok {
		return open
	}
	return true
}
