package templates

import (
	"context"
	"net/http"
	"strings"
)

// CurrentPathMiddleware stores the request path in the context so server-rendered
// templates can mark the active sidebar link with aria-current="page".
// basecoat-css 0.3.11 removed the client-side highlighting, so Annict applies
// it on the server instead. It lives in this package (rather than internal/middleware)
// to avoid an import cycle, since middleware is already imported by i18n which
// this package depends on.
//
// [Ja] CurrentPathMiddleware はリクエストパスをコンテキストに保存し、サーバーレンダリング
// のテンプレートが現在ページのサイドバーリンクに aria-current="page" を付与できるように
// する。basecoat-css 0.3.11 でクライアントサイドのハイライトが削除されたため、Annict では
// サーバー側で付与する。internal/middleware ではなくこのパッケージに置くのは、middleware が
// 既に (このパッケージが依存する) i18n から import されており、インポート循環を避けるため。
func CurrentPathMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := SetCurrentPath(r.Context(), r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// currentPathContextKey is the context key under which the current request
// path is stored, so templates can mark the active link with aria-current.
//
// [Ja] currentPathContextKey は現在のリクエストパスを保存するコンテキストキー。
// テンプレートが現在ページのリンクに aria-current を付与するために使う。
type currentPathContextKey struct{}

// SetCurrentPath stores the current request path in the context.
//
// [Ja] SetCurrentPath は現在のリクエストパスをコンテキストに保存する。
func SetCurrentPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, currentPathContextKey{}, path)
}

// GetCurrentPath returns the current request path stored in the context.
// Returns an empty string when no path has been set.
//
// [Ja] GetCurrentPath はコンテキストに保存された現在のリクエストパスを返す。
// パスが設定されていない場合は空文字列を返す。
func GetCurrentPath(ctx context.Context) string {
	if path, ok := ctx.Value(currentPathContextKey{}).(string); ok {
		return path
	}
	return ""
}

// IsCurrentPath reports whether the given link path matches the current page.
// Both paths are normalized (query/fragment stripped, trailing slash trimmed)
// so that, for example, "/track" and "/track/" are treated as the same page.
// This replaces the client-side highlighting that basecoat-css removed in
// 0.3.11.
//
// [Ja] IsCurrentPath は与えられたリンクパスが現在ページと一致するかを返す。
// 両者のパスを正規化 (クエリ/フラグメント除去・末尾スラッシュ除去) して比較するため、
// 例えば "/track" と "/track/" は同一ページとして扱う。basecoat-css 0.3.11 で
// 削除されたクライアントサイドのハイライトを置き換えるもの。
func IsCurrentPath(ctx context.Context, path string) bool {
	return normalizePath(GetCurrentPath(ctx)) == normalizePath(path)
}

// normalizePath strips any query/fragment and the trailing slash (except for
// the root path) so that equivalent paths compare equal.
//
// [Ja] normalizePath はクエリ/フラグメントと (ルート以外の) 末尾スラッシュを除去し、
// 等価なパス同士が一致するようにする。
func normalizePath(p string) string {
	if i := strings.IndexAny(p, "?#"); i >= 0 {
		p = p[:i]
	}
	if len(p) > 1 {
		p = strings.TrimRight(p, "/")
	}
	return p
}
