// Package templates はHTMLテンプレート機能を提供します
package templates

import (
	"context"
	"encoding/json"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/i18n"
)

// ========================================
// templ用ヘルパー関数
// ========================================

// T は翻訳を取得する（templ用）
func T(ctx context.Context, messageID string, data ...map[string]any) string {
	return i18n.T(ctx, messageID, data...)
}

// Locale は現在のロケールを取得する
func Locale(ctx context.Context) string {
	return i18n.GetLocale(ctx)
}

// HXCSRFHeaders returns the JSON value for htmx's hx-headers attribute that sends the CSRF
// token in the X-CSRF-Token header. htmx-issued DELETE/POST requests carry no parseable
// form body (net/http only parses the body of POST/PUT/PATCH), so the CSRF middleware reads
// the token from this header instead.
//
// [Ja] HXCSRFHeaders は htmx の hx-headers 属性に渡す JSON を返す。CSRF トークンを
// X-CSRF-Token ヘッダーで送るためのもの。htmx が発行する DELETE/POST リクエストには
// パース可能なフォーム本体が無い (net/http は POST/PUT/PATCH の本体しかパースしない) ため、
// CSRF ミドルウェアはこのヘッダーからトークンを読む。
func HXCSRFHeaders(token string) string {
	b, err := json.Marshal(map[string]string{"X-CSRF-Token": token})
	if err != nil {
		return "{}"
	}
	return string(b)
}

// Deref はポインタを参照外しする（ジェネリック対応）
func Deref[T any](v *T) T {
	if v != nil {
		return *v
	}
	var zero T
	return zero
}

// Icon はアイコン名からSVGを返す（templ.Component対応）
// 可変長引数でクラス名を指定可能: Icon("name", "class1 class2")
func Icon(name string, class ...string) templ.Component {
	svg, ok := phosphorIcons[name]
	if !ok {
		// デフォルトとしてinfoアイコンを使用
		svg = phosphorIcons["info"]
	}

	// クラス名が指定されている場合は、SVGタグに追加
	if len(class) > 0 && class[0] != "" {
		// <svg の直後にclass属性を挿入
		svg = `<svg class="` + class[0] + `" ` + svg[5:]
	}

	// templ.Rawを使用してSVGを返す
	return templ.Raw(svg)
}
