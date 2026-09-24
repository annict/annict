// Package templatesはHTMLテンプレート機能を提供します
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

// Tは翻訳を取得する (templ用)
func T(ctx context.Context, messageID string, data ...map[string]any) string {
	return i18n.T(ctx, messageID, data...)
}

// Localeは現在のロケールを取得する
func Locale(ctx context.Context) string {
	return i18n.GetLocale(ctx)
}

// HXCSRFHeadersはhtmxのhx-headers属性に渡すJSONを返す。CSRFトークンを
// X-CSRF-Tokenヘッダーで送るためのもの。htmxが発行するDELETE/POSTリクエストには
// パース可能なフォーム本体が無い (net/httpはPOST/PUT/PATCHの本体しかパースしない) ため、
// CSRFミドルウェアはこのヘッダーからトークンを読む。
func HXCSRFHeaders(token string) string {
	b, err := json.Marshal(map[string]string{"X-CSRF-Token": token})
	if err != nil {
		return "{}"
	}
	return string(b)
}

// Derefはポインタを参照外しする (ジェネリック対応)
func Deref[T any](v *T) T {
	if v != nil {
		return *v
	}
	var zero T
	return zero
}

// Iconは指定したアイコン名のSVGコンポーネントを、アクセシビリティ属性を追加せずに
// 返す。省略可能なclass引数はSVG要素に追加する。SVG自体が意味を伝える場合は
// LabeledIconを、近くのテキストを繰り返すだけの場合はDecorativeIconを使う。
func Icon(name string, class ...string) templ.Component {
	return templ.Raw(iconSVG(name, class...))
}

// LabeledIconは指定したアクセシブルネームを持つ画像として公開するSVGを返す。
// labelには空でない、現在のページの言語に翻訳済みの文字列を渡す。近くのテキストだけでは
// 伝わらない情報をSVG自体が担う場合にのみ使う。
func LabeledIcon(name, label string, class ...string) templ.Component {
	svg := iconSVG(name, class...)
	return templ.Raw(`<svg role="img" aria-label="` + templ.EscapeString(label) + `" focusable="false" ` + svg[5:])
}

// InlineIconPositionはボタンのテキストに添えるアイコンについて、Basecoatが受け付ける
// 位置を表す。DecorativeInlineIconはdata-icon属性として出力する前に値を検証する。
type InlineIconPosition string

const (
	InlineIconStart InlineIconPosition = "inline-start"
	InlineIconEnd   InlineIconPosition = "inline-end"
)

// DecorativeIconは支援技術から隠し、フォーカス順序から除外したSVGコンポーネントを
// 返す。Basecoatのテキスト付きボタン以外で、近くのテキストを繰り返すアイコンに使う。
// ボタンのテキストに添える場合はDecorativeInlineIconを使う。それ自体が意味を伝えるSVGは
// LabeledIconが担当する。
func DecorativeIcon(name string, class ...string) templ.Component {
	svg := iconSVG(name, class...)
	return templ.Raw(`<svg aria-hidden="true" focusable="false" ` + svg[5:])
}

// DecorativeInlineIconはBasecoatの検証済みinline位置を持つ装飾SVGを返す。不明な
// 位置はマークアップへコピーせず省略し、アイコンは引き続き支援技術から隠してフォーカス順序
// から除外する。
func DecorativeInlineIcon(name string, position InlineIconPosition, class ...string) templ.Component {
	switch position {
	case InlineIconStart, InlineIconEnd:
		svg := iconSVG(name, class...)
		return templ.Raw(`<svg data-icon="` + string(position) + `" aria-hidden="true" focusable="false" ` + svg[5:])
	default:
		return DecorativeIcon(name, class...)
	}
}

// fallbackIconNameはphosphorIconsが持たない名前の代わりに描画するアイコン。名前を
// 間違えても何も出ないのではなく、目に見える代替が出るようにするため。mapはこのエントリを
// 持つ必要がある。iconSVGは解決したマークアップ先頭の "<svg " を置き換えて結果を組み立てる
// ため、エントリが無いと置き換える前置き自体が無くなる。この不変条件は
// TestPhosphorIconsHoldFallbackIconが担保する。
const fallbackIconName = "info"

// iconSVGは1つのアイコンの保持しているマークアップに、省略可能なclassを適用して
// 返す。Icon・LabeledIcon・DecorativeIcon・DecorativeInlineIconはいずれもこの戻り値を描画
// するため、あるアイコン名がどのマークアップになるかで食い違うことはない。
//
// 属性は保持しているマークアップ先頭の "<svg " を置き換えて足すため、phosphorIconsの各
// エントリはちょうどこの5文字で始まる必要がある。この不変条件は
// TestPhosphorIconsStartWithSVGTagが担保する。別の書き方のアイコンは、エラーになるのでは
// なく壊れたマークアップを生むため。
func iconSVG(name string, class ...string) string {
	svg, ok := phosphorIcons[name]
	if !ok {
		svg = phosphorIcons[fallbackIconName]
	}

	if len(class) > 0 && class[0] != "" {
		svg = `<svg class="` + class[0] + `" ` + svg[5:]
	}

	return svg
}
