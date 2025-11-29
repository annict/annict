package testutil

import (
	"net/http"
	"testing"

	"github.com/annict/annict/internal/i18n"
)

// ApplyI18nMiddleware はテスト用にI18nミドルウェアを適用するヘルパー関数
// テストでハンドラーを実行する際に、リクエストのコンテキストにlocaleを設定する
func ApplyI18nMiddleware(t *testing.T, handler http.HandlerFunc) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		// I18nミドルウェアを通してからハンドラーを実行
		i18n.Middleware(handler).ServeHTTP(w, r)
	}
}
