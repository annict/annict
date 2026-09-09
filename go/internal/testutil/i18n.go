package testutil

import (
	"net/http"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

// ApplyI18nMiddleware runs handler with the locale the i18n middleware would have resolved for
// an unauthenticated request, so that tests exercise handlers with a locale on the context.
//
// It reproduces that resolution instead of calling middleware.I18n: internal/middleware renders
// shared error pages through internal/httperror, so importing it here would put internal/templates
// and every package below it into an import cycle with their own in-package tests, which depend on
// this package.
//
// [Ja] ApplyI18nMiddleware は、未認証リクエストに対して i18n ミドルウェアが解決したはずの
// ロケールを与えて handler を実行する。テストがロケールの載ったコンテキストでハンドラーを
// 動かせるようにするため。
//
// middleware.I18n を呼ばずに解決処理を再現しているのは、internal/middleware が共通エラーページを
// internal/httperror 経由で描画するため。ここで import すると、本パッケージに依存する
// internal/templates 以下の各パッケージの内部テストが import 循環になる。
func ApplyI18nMiddleware(t *testing.T, handler http.HandlerFunc) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := i18n.SetLocale(r.Context(), i18n.DetectLanguage(r))

		handler.ServeHTTP(w, r.WithContext(ctx))
	}
}
