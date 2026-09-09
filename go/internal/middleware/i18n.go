package middleware

import (
	"net/http"

	"github.com/annict/annict/go/internal/i18n"
)

// I18n stores the request's locale on the context: users.locale for a signed-in user, and the
// Accept-Language header otherwise.
//
// It lives here rather than in internal/i18n because resolving the locale needs the signed-in
// user from this package's context. Keeping it there would make internal/i18n depend on this
// package, and this package could then render nothing that is translated
// (middleware → httperror → i18n → middleware).
//
// [Ja] I18n はリクエストのロケールをコンテキストに保存する。ログイン済みユーザーの場合は
// users.locale を、未ログインの場合は Accept-Language ヘッダーを使う。
//
// 本ミドルウェアが internal/i18n ではなくここにあるのは、ロケールの解決に本パッケージの
// コンテキストが持つログイン中のユーザーが要るため。あちらに置くと internal/i18n が本パッケージへ
// 依存し、本パッケージは翻訳を伴う描画を一切行えなくなる
// (middleware → httperror → i18n → middleware の循環)。
func I18n(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := i18n.SetLocale(r.Context(), resolveLocale(r))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// resolveLocale returns the locale the request should be rendered in. A signed-in user's
// preference wins over the header, falling back to the default language when the stored value
// is not one this application ships translations for.
//
// [Ja] resolveLocale はリクエストを描画すべきロケールを返す。ログイン済みユーザーの設定が
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
