package viewmodel

import "github.com/annict/annict/go/internal/model"

// FormErrors is the Presentation-layer view of what a submitted form got back: the messages
// that belong to the form as a whole, and the messages each field collected. Templates render
// a rejected submit from this type instead of from *model.ValidationError, which keeps them
// off the domain (the templates-layer rule in .golangci.yml denies internal/model).
//
// The shape mirrors *model.ValidationError rather than reworking it, because the pages already
// render errors exactly the way the domain groups them: global messages at the top, field
// messages beside their input. A different shape here would be a second grouping to keep in
// step with the first for no gain at the screen.
//
// [Ja] FormErrors は、送信されたフォームが受け取ったもの (フォーム全体に属するメッセージと、
// 各フィールドが集めたメッセージ) を表す Presentation 層の型。テンプレートは却下された送信を
// *model.ValidationError ではなくこの型から描画し、ドメインから切り離される (.golangci.yml の
// templates-layer が internal/model を deny している)。
//
// 形は *model.ValidationError を作り直さず写している。各ページは既にドメインの区分どおりに
// エラーを描画しており (グローバルなメッセージは冒頭、フィールドのメッセージはその入力欄の
// 傍ら)、ここで別の区分を作っても画面上の利点が無いまま、揃え続ける対象が 2 つに増えるため。
type FormErrors struct {
	// Global holds the messages that belong to the form as a whole rather than to one of its
	// fields.
	//
	// [Ja] Global は個々のフィールドではなくフォーム全体に属するメッセージを保持する。
	Global []string
	// Fields holds the messages each field collected, keyed by the field's name.
	//
	// [Ja] Fields は各フィールドが集めたメッセージを、そのフィールド名をキーとして保持する。
	Fields map[string][]string
}

// NewFormErrors projects a validation error onto the Presentation layer. It returns nil for a
// nil error, so a page that was not submitted and a submit that was accepted both reach the
// template with no errors, exactly as they did while the template took the domain type.
//
// The projection shares the underlying slices and map with ve instead of copying them. Its
// purpose is the type boundary, not isolation from later writes: the handler builds it from an
// error it has just caught and renders, and neither side is written to afterwards.
//
// [Ja] NewFormErrors はバリデーションエラーを Presentation 層へ射影する。nil のエラーには nil
// を返すため、送信されていないページと受け付けられた送信は、テンプレートがドメインの型を受け
// 取っていたときと同じく、いずれもエラー無しでテンプレートに届く。
//
// 射影は ve のスライスと map を複製せず共有する。目的は型の境界であって、後続の書き込みからの
// 隔離ではない。ハンドラーは捕捉した直後のエラーからこれを組み立てて描画し、その後どちらも書か
// れないため。
func NewFormErrors(ve *model.ValidationError) *FormErrors {
	if ve == nil {
		return nil
	}

	return &FormErrors{
		Global: ve.Global,
		Fields: ve.Fields,
	}
}

// HasErrors reports whether the submit collected any message at all.
//
// [Ja] HasErrors は送信が 1 件でもメッセージを集めたかどうかを返す。
func (e *FormErrors) HasErrors() bool {
	if e == nil {
		return false
	}

	return len(e.Global) > 0 || len(e.Fields) > 0
}

// HasFieldError reports whether the named field collected a message.
//
// [Ja] HasFieldError は名前で指定したフィールドがメッセージを集めたかどうかを返す。
func (e *FormErrors) HasFieldError(field string) bool {
	if e == nil {
		return false
	}

	return len(e.Fields[field]) > 0
}

// GetFieldErrors returns the messages the named field collected.
//
// [Ja] GetFieldErrors は名前で指定したフィールドが集めたメッセージを返す。
func (e *FormErrors) GetFieldErrors(field string) []string {
	if e == nil {
		return nil
	}

	return e.Fields[field]
}
