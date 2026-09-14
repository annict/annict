package middleware

import (
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
)

// I18nはリクエストのロケールをコンテキストに保存する。ログイン済みユーザーの場合は
// users.localeを、未ログインの場合はAccept-Languageヘッダーを使う。
//
// 本ミドルウェアがinternal/i18nではなくここにあるのは、ロケールの解決に本パッケージの
// コンテキストが持つログイン中のユーザーが要るため。あちらに置くとinternal/i18nが本パッケージへ
// 依存し、本パッケージは翻訳を伴う描画を一切行えなくなる
// (middleware → httperror → i18n → middlewareの循環)。
func I18n(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := i18n.SetLocale(r.Context(), resolveLocale(r))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// resolveLocaleはリクエストを描画すべきロケールを返す。ログイン済みユーザーの設定が
// ヘッダーより優先され、保存されている値が本アプリケーションの翻訳を持たない言語だった場合は
// デフォルト言語にフォールバックする。
func resolveLocale(r *http.Request) string {
	user := GetUserFromContext(r.Context())
	if user == nil {
		return i18n.DetectLanguage(r)
	}

	if !i18n.IsSupportedLang(user.Locale) {
		return i18n.DefaultLang
	}

	return user.Locale
}
