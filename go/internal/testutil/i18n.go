package testutil

import (
	"net/http"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

// ApplyI18nMiddlewareは、未認証リクエストに対してi18nミドルウェアが解決したはずの
// ロケールを与えてhandlerを実行する。テストがロケールの載ったコンテキストでハンドラーを
// 動かせるようにするため。
//
// middleware.I18nを呼ばずに解決処理を再現しているのは、internal/middlewareが共通エラーページを
// internal/httperror経由で描画するため。ここでimportすると、本パッケージに依存する
// internal/templates以下の各パッケージの内部テストがimport循環になる。
func ApplyI18nMiddleware(t *testing.T, handler http.HandlerFunc) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := i18n.SetLocale(r.Context(), i18n.DetectLanguage(r))

		handler.ServeHTTP(w, r.WithContext(ctx))
	}
}
