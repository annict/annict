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

// Icon returns an SVG component for the given icon name without adding accessibility
// attributes. The optional class argument is added to the SVG element. Use LabeledIcon when
// the SVG itself conveys meaning, or DecorativeIcon when it only repeats nearby text.
//
// [Ja] Icon は指定したアイコン名の SVG コンポーネントを、アクセシビリティ属性を追加せずに
// 返す。省略可能な class 引数は SVG 要素に追加する。SVG 自体が意味を伝える場合は
// LabeledIcon を、近くのテキストを繰り返すだけの場合は DecorativeIcon を使う。
func Icon(name string, class ...string) templ.Component {
	return templ.Raw(iconSVG(name, class...))
}

// LabeledIcon returns an SVG exposed as an image with the given accessible label. The label
// must be non-empty and localized for the current page. Use it only when the SVG itself
// conveys information that nearby text does not already provide.
//
// [Ja] LabeledIcon は指定したアクセシブルネームを持つ画像として公開する SVG を返す。
// label には空でない、現在のページの言語に翻訳済みの文字列を渡す。近くのテキストだけでは
// 伝わらない情報を SVG 自体が担う場合にのみ使う。
func LabeledIcon(name, label string, class ...string) templ.Component {
	svg := iconSVG(name, class...)
	return templ.Raw(`<svg role="img" aria-label="` + templ.EscapeString(label) + `" focusable="false" ` + svg[5:])
}

// InlineIconPosition names the placement Basecoat accepts for an icon beside button text.
// DecorativeInlineIcon validates the value before exposing it as a data-icon attribute.
//
// [Ja] InlineIconPosition はボタンのテキストに添えるアイコンについて、Basecoat が受け付ける
// 位置を表す。DecorativeInlineIcon は data-icon 属性として出力する前に値を検証する。
type InlineIconPosition string

const (
	InlineIconStart InlineIconPosition = "inline-start"
	InlineIconEnd   InlineIconPosition = "inline-end"
)

// DecorativeIcon returns an SVG component hidden from assistive technology and removed from
// the focus order. Use it for an icon that repeats nearby text outside a Basecoat text button;
// use DecorativeInlineIcon when the icon sits beside a button's text. An SVG that conveys
// meaning of its own belongs in LabeledIcon.
//
// [Ja] DecorativeIcon は支援技術から隠し、フォーカス順序から除外した SVG コンポーネントを
// 返す。Basecoat のテキスト付きボタン以外で、近くのテキストを繰り返すアイコンに使う。
// ボタンのテキストに添える場合は DecorativeInlineIcon を使う。それ自体が意味を伝える SVG は
// LabeledIcon が担当する。
func DecorativeIcon(name string, class ...string) templ.Component {
	svg := iconSVG(name, class...)
	return templ.Raw(`<svg aria-hidden="true" focusable="false" ` + svg[5:])
}

// DecorativeInlineIcon returns a decorative SVG with Basecoat's validated inline position.
// Unknown positions are omitted instead of being copied into markup, while the icon remains
// hidden from assistive technology and excluded from the focus order.
//
// [Ja] DecorativeInlineIcon は Basecoat の検証済み inline 位置を持つ装飾 SVG を返す。不明な
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

// fallbackIconName is drawn in place of a name phosphorIcons does not hold, so a mistyped
// name shows a visible placeholder instead of nothing. The map has to hold it: iconSVG builds
// its result by replacing the leading "<svg " of what it resolved, and a missing entry leaves
// no such prefix to replace. TestPhosphorIconsHoldFallbackIcon holds that invariant.
//
// [Ja] fallbackIconName は phosphorIcons が持たない名前の代わりに描画するアイコン。名前を
// 間違えても何も出ないのではなく、目に見える代替が出るようにするため。map はこのエントリを
// 持つ必要がある。iconSVG は解決したマークアップ先頭の "<svg " を置き換えて結果を組み立てる
// ため、エントリが無いと置き換える前置き自体が無くなる。この不変条件は
// TestPhosphorIconsHoldFallbackIcon が担保する。
const fallbackIconName = "info"

// iconSVG returns the stored markup of one icon with the optional class applied. Icon,
// LabeledIcon, DecorativeIcon, and DecorativeInlineIcon all render what it returns, so they
// cannot disagree about which markup an icon name resolves to.
//
// Attributes are added by replacing the leading "<svg " of the stored markup, so every entry
// of phosphorIcons has to start with exactly those five characters.
// TestPhosphorIconsStartWithSVGTag holds that invariant: an icon written any other way would
// produce broken markup rather than fail outright.
//
// [Ja] iconSVG は 1 つのアイコンの保持しているマークアップに、省略可能な class を適用して
// 返す。Icon・LabeledIcon・DecorativeIcon・DecorativeInlineIcon はいずれもこの戻り値を描画
// するため、あるアイコン名がどのマークアップになるかで食い違うことはない。
//
// 属性は保持しているマークアップ先頭の "<svg " を置き換えて足すため、phosphorIcons の各
// エントリはちょうどこの 5 文字で始まる必要がある。この不変条件は
// TestPhosphorIconsStartWithSVGTag が担保する。別の書き方のアイコンは、エラーになるのでは
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
