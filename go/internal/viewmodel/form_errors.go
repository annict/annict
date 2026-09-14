package viewmodel

import "github.com/annict/annict/go/internal/model"

// FormErrorsは、送信されたフォームが受け取ったもの (フォーム全体に属するメッセージと、
// 各フィールドが集めたメッセージ) を表すPresentation層の型。テンプレートは却下された送信を
// *model.ValidationErrorではなくこの型から描画し、ドメインから切り離される (.golangci.ymlの
// templates-layerがinternal/modelをdenyしている)。
//
// 形は *model.ValidationErrorを作り直さず写している。各ページは既にドメインの区分どおりに
// エラーを描画しており (グローバルなメッセージは冒頭、フィールドのメッセージはその入力欄の
// 傍ら)、ここで別の区分を作っても画面上の利点が無いまま、揃え続ける対象が2つに増えるため。
type FormErrors struct {
	// Globalは個々のフィールドではなくフォーム全体に属するメッセージを保持する。
	Global []string
	// Fieldsは各フィールドが集めたメッセージを、そのフィールド名をキーとして保持する。
	Fields map[string][]string
}

// NewFormErrorsはバリデーションエラーをPresentation層へ射影する。nilのエラーにはnil
// を返すため、送信されていないページと受け付けられた送信は、テンプレートがドメインの型を受け
// 取っていたときと同じく、いずれもエラー無しでテンプレートに届く。
//
// 射影はveのスライスとmapを複製せず共有する。目的は型の境界であって、後続の書き込みからの
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

// HasErrorsは送信が1件でもメッセージを集めたかどうかを返す。
func (e *FormErrors) HasErrors() bool {
	if e == nil {
		return false
	}

	return len(e.Global) > 0 || len(e.Fields) > 0
}

// HasFieldErrorは名前で指定したフィールドがメッセージを集めたかどうかを返す。
func (e *FormErrors) HasFieldError(field string) bool {
	if e == nil {
		return false
	}

	return len(e.Fields[field]) > 0
}

// GetFieldErrorsは名前で指定したフィールドが集めたメッセージを返す。
func (e *FormErrors) GetFieldErrors(field string) []string {
	if e == nil {
		return nil
	}

	return e.Fields[field]
}
