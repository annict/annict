package templates

import (
	"context"
	"net/http"
	"strings"
)

// CurrentPathMiddlewareはリクエストパスをコンテキストに保存し、サーバーレンダリング
// のテンプレートが現在ページのサイドバーリンクにaria-current="page" を付与できるように
// する。basecoat-css 0.3.11でクライアントサイドのハイライトが削除されたため、Annictでは
// サーバー側で付与する。internal/middlewareではなくこのパッケージに置くのは、middlewareが
// 既に (このパッケージが依存する) i18nからimportされており、インポート循環を避けるため。
func CurrentPathMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := SetCurrentPath(r.Context(), r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// currentPathContextKeyは現在のリクエストパスを保存するコンテキストキー。
// テンプレートが現在ページのリンクにaria-currentを付与するために使う。
type currentPathContextKey struct{}

// SetCurrentPathは現在のリクエストパスをコンテキストに保存する。
func SetCurrentPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, currentPathContextKey{}, path)
}

// GetCurrentPathはコンテキストに保存された現在のリクエストパスを返す。
// パスが設定されていない場合は空文字列を返す。
func GetCurrentPath(ctx context.Context) string {
	if path, ok := ctx.Value(currentPathContextKey{}).(string); ok {
		return path
	}
	return ""
}

// IsCurrentPathは与えられたリンクパスが現在ページと一致するかを返す。
// 両者のパスを正規化 (クエリ/フラグメント除去・末尾スラッシュ除去) して比較するため、
// 例えば "/track" と "/track/" は同一ページとして扱う。basecoat-css 0.3.11で
// 削除されたクライアントサイドのハイライトを置き換えるもの。
func IsCurrentPath(ctx context.Context, path string) bool {
	return normalizePath(GetCurrentPath(ctx)) == normalizePath(path)
}

// IsCurrentPathPrefixは現在ページが与えられたリンク先そのものか、その配下のページかを
// 返す。ナビゲーションの項目が、その画面群のどこにいる間も印を保てるようにするためのもの
// (例: "/db/works/1/edit" を開いている間の "/db/works")。比較はセグメント境界でのみ一致と
// みなすため、"/db/series_works/1" は "/db/series" に一致しない。そのページ自身にだけ印を
// 付けたい場合はIsCurrentPathを使う。
func IsCurrentPathPrefix(ctx context.Context, path string) bool {
	current := normalizePath(GetCurrentPath(ctx))
	link := normalizePath(path)

	return current == link || strings.HasPrefix(current, link+"/")
}

// normalizePathはクエリ/フラグメントと (ルート以外の) 末尾スラッシュを除去し、
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
