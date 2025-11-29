// Package i18n はi18n機能を提供します
package i18n

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	authMiddleware "github.com/annict/annict/internal/middleware"
)

// 翻訳ファイルを埋め込み
//
//go:embed locales/*.toml
var localesFS embed.FS

// サポートする言語
const (
	LangJa      = "ja"
	LangEn      = "en"
	DefaultLang = LangJa
)

// contextキーの型
type contextKey string

const (
	localeContextKey    contextKey = "locale"
	localizerContextKey contextKey = "localizer"
)

// グローバルなバンドル
var bundle *i18n.Bundle

// init でlocalesディレクトリから全ての翻訳ファイルを読み込む
func init() {
	// 日本語をデフォルト言語として設定
	bundle = i18n.NewBundle(language.Japanese)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// 翻訳ファイルを読み込み
	languages := []struct {
		code string
		tag  language.Tag
	}{
		{LangJa, language.Japanese},
		{LangEn, language.English},
	}

	for _, lang := range languages {
		data, err := localesFS.ReadFile(fmt.Sprintf("locales/%s.toml", lang.code))
		if err != nil {
			continue
		}

		bundle.MustParseMessageFileBytes(data, fmt.Sprintf("%s.toml", lang.code))
	}
}

// T は翻訳関数（テンプレートから呼び出される）
func T(ctx context.Context, messageID string, templateData ...map[string]any) string {
	localizer := GetLocalizer(ctx)
	if localizer == nil {
		return messageID
	}

	config := &i18n.LocalizeConfig{
		MessageID: messageID,
	}

	// テンプレートデータがある場合は設定
	if len(templateData) > 0 && templateData[0] != nil {
		config.TemplateData = templateData[0]

		// Countが含まれている場合は複数形処理を有効にする
		if count, ok := templateData[0]["Count"].(int32); ok {
			config.PluralCount = int(count)
		} else if count, ok := templateData[0]["Count"].(int); ok {
			config.PluralCount = count
		}
	}

	message, err := localizer.Localize(config)
	if err != nil {
		// 翻訳が見つからない場合はメッセージIDを返す
		return messageID
	}

	return message
}

// GetLocale はコンテキストから言語設定を取得する
func GetLocale(ctx context.Context) string {
	if locale, ok := ctx.Value(localeContextKey).(string); ok {
		return locale
	}
	return DefaultLang
}

// SetLocale はコンテキストに言語設定を保存する
func SetLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeContextKey, locale)
}

// GetLocalizer はコンテキストからLocalizerを取得する
func GetLocalizer(ctx context.Context) *i18n.Localizer {
	if localizer, ok := ctx.Value(localizerContextKey).(*i18n.Localizer); ok {
		return localizer
	}
	// Localizerがない場合は作成
	locale := GetLocale(ctx)
	return i18n.NewLocalizer(bundle, locale)
}

// SetLocalizer はコンテキストにLocalizerを保存する
func SetLocalizer(ctx context.Context, localizer *i18n.Localizer) context.Context {
	return context.WithValue(ctx, localizerContextKey, localizer)
}

// DetectLanguage はリクエストのAccept-Languageヘッダーから言語を検出する
func DetectLanguage(r *http.Request) string {
	// Accept-Languageヘッダーから取得
	acceptLang := r.Header.Get("Accept-Language")
	if strings.Contains(acceptLang, "ja") {
		return LangJa
	}
	// jaが含まれていない場合のみenをチェック
	if strings.Contains(acceptLang, "en") {
		return LangEn
	}

	// デフォルトは日本語
	return DefaultLang
}

// Middleware はI18nミドルウェアを提供する
// ログイン済みユーザーの場合はusers.localeを、未ログインの場合はAccept-Languageヘッダーから言語を決定する
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var locale string

		// コンテキストからユーザー情報を取得
		user := authMiddleware.GetUserFromContext(r.Context())
		if user != nil {
			// ログイン済み: users.localeを使用
			locale = user.Locale
			// 念のため、サポート対象の言語かチェック
			if locale != LangJa && locale != LangEn {
				locale = DefaultLang
			}
		} else {
			// 未ログイン: Accept-Languageヘッダーから検出
			locale = DetectLanguage(r)
		}

		// Localizerを作成
		localizer := i18n.NewLocalizer(bundle, locale)

		// コンテキストに設定
		ctx := SetLocale(r.Context(), locale)
		ctx = SetLocalizer(ctx, localizer)

		// 次のハンドラーへ
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
