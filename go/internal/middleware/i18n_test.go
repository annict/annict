package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/middleware"
	"github.com/annict/annict/go/internal/model"
)

// TestI18n verifies which locale reaches the handler: a signed-in user's preference outranks
// the header, an unsupported stored value falls back to the default language, and a request
// without a user follows Accept-Language.
//
// [Ja] TestI18n はハンドラーにどのロケールが届くかを検証する。ログイン済みユーザーの設定が
// ヘッダーより優先されること、保存値が未対応の言語ならデフォルト言語にフォールバックすること、
// ユーザーの無いリクエストは Accept-Language に従うことを確認する。
func TestI18n(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		acceptLanguage string
		user           *model.User
		wantLocale     string
	}{
		{
			name:           "未認証はAccept-Languageに従う",
			acceptLanguage: "en-US,en;q=0.9",
			user:           nil,
			wantLocale:     i18n.LangEn,
		},
		{
			name:           "未認証でAccept-Languageが未対応の言語ならデフォルト言語",
			acceptLanguage: "fr",
			user:           nil,
			wantLocale:     i18n.DefaultLang,
		},
		{
			name:           "ログイン済みはusers.localeがAccept-Languageより優先される",
			acceptLanguage: "ja",
			user:           newUserWithLocale(i18n.LangEn),
			wantLocale:     i18n.LangEn,
		},
		{
			name:           "ログイン済みでusers.localeが未対応の言語ならデフォルト言語",
			acceptLanguage: "en",
			user:           newUserWithLocale("fr"),
			wantLocale:     i18n.DefaultLang,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotLocale string
			var gotTranslation string
			handler := middleware.I18n(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotLocale = i18n.GetLocale(r.Context())
				gotTranslation = i18n.T(r.Context(), "error_not_found_title")
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Language", tt.acceptLanguage)
			if tt.user != nil {
				req = setUserContext(req, tt.user)
			}

			handler.ServeHTTP(httptest.NewRecorder(), req)

			if gotLocale != tt.wantLocale {
				t.Errorf("locale = %q, want %q", gotLocale, tt.wantLocale)
			}

			// The localizer travels with the locale, so a translation resolved downstream
			// is in the same language rather than the message ID.
			//
			// [Ja] Localizer はロケールと一緒に運ばれるため、下流で解決した翻訳は
			// メッセージ ID ではなく同じ言語の文言になる。
			wantTranslation := "ページが見つかりません"
			if tt.wantLocale == i18n.LangEn {
				wantTranslation = "Page not found"
			}
			if gotTranslation != wantTranslation {
				t.Errorf("翻訳 = %q, want %q", gotTranslation, wantTranslation)
			}
		})
	}
}

func newUserWithLocale(locale string) *model.User {
	user := newUserWithRole(middleware.RoleUser)
	user.Locale = locale

	return user
}
