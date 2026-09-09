// Package i18n provides translation lookup and the request locale carried on the context.
//
// It depends on no other package of this application: resolving which locale a request
// belongs to needs the signed-in user, and that lives in internal/middleware, so keeping the
// resolution here would make internal/middleware unable to render anything this package
// translates (middleware → httperror → i18n → middleware). The resolution is
// middleware.I18n instead, and this package stays a leaf that only translates.
//
// [Ja] Package i18n は翻訳の取得と、コンテキストが運ぶリクエストのロケールを提供する。
//
// 本パッケージはアプリケーション内の他のパッケージに依存しない。リクエストのロケールの解決には
// ログイン中のユーザーが要り、それは internal/middleware にあるため、解決処理をここに置くと
// internal/middleware が本パッケージの翻訳を使う描画を行えなくなる
// (middleware → httperror → i18n → middleware の循環)。解決は middleware.I18n が担い、
// 本パッケージは翻訳だけを担う leaf に保つ。
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

// SetLocale stores the locale on the context together with a Localizer built for it, so that
// the translations a request renders are resolved once instead of per T call.
//
// [Ja] SetLocale はコンテキストに言語設定を保存し、あわせてその言語の Localizer も保存する。
// 1 リクエストが描画する翻訳の解決を T の呼び出しごとではなく 1 回で済ませるため。
func SetLocale(ctx context.Context, locale string) context.Context {
	ctx = context.WithValue(ctx, localeContextKey, locale)
	return context.WithValue(ctx, localizerContextKey, i18n.NewLocalizer(bundle, locale))
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

// IsSupportedLang reports whether locale is one of the languages this application ships
// translations for.
//
// [Ja] IsSupportedLang は locale が本アプリケーションが翻訳を持つ言語かどうかを返す。
func IsSupportedLang(locale string) bool {
	return locale == LangJa || locale == LangEn
}
