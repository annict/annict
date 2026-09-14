package viewmodel

import (
	"slices"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// TestNewFormErrors_NilはnilのバリデーションエラーがnilのFormErrorsへ射影される
// ことを検証する。ハンドラーは射影をそのままページのデータへ渡し、各フォームのテンプレートは
// nilを「この送信は却下されていない」として扱うため、空の非nilを返すと誰も送信していない
// フォームのエラー要約が描画されてしまう。
func TestNewFormErrors_Nil(t *testing.T) {
	t.Parallel()

	if got := NewFormErrors(nil); got != nil {
		t.Errorf("NewFormErrors(nil) = %v、期待値 = nil", got)
	}
}

// TestNewFormErrors_MirrorsValidationErrorは、射影が元のドメインのエラーと同じメッセージ
// を返すことを検証する。テンプレートは読み方を変えないまま *model.ValidationErrorから本型へ
// 移ったため、メッセージを落としたり並べ替えたりする射影は、どのテンプレートも変わらないまま
// 却下された送信の表示を変えてしまう。
func TestNewFormErrors_MirrorsValidationError(t *testing.T) {
	t.Parallel()

	ve := model.NewValidationError()
	ve.AddGlobal("フォーム全体のエラー")
	ve.AddField("title", "入力してください")
	ve.AddField("title", "50文字以内で入力してください")
	ve.AddField("number", "数値で入力してください")

	got := NewFormErrors(ve)

	if !slices.Equal(got.Global, ve.Global) {
		t.Errorf("Global = %v、期待値 = %v", got.Global, ve.Global)
	}

	for _, field := range []string{"title", "number", "absent"} {
		if want := ve.HasFieldError(field); got.HasFieldError(field) != want {
			t.Errorf("HasFieldError(%q) = %v、期待値 = %v", field, got.HasFieldError(field), want)
		}
		if want := ve.GetFieldErrors(field); !slices.Equal(got.GetFieldErrors(field), want) {
			t.Errorf("GetFieldErrors(%q) = %v、期待値 = %v", field, got.GetFieldErrors(field), want)
		}
	}

	if !got.HasErrors() {
		t.Error("HasErrors() = false、期待値 = true")
	}
}

// TestFormErrors_HasErrorsはフォームのテンプレートが区別する3つの状態 (射影が無い、
// グローバルメッセージだけを持つ、フィールドのメッセージだけを持つ) を扱う。各ページはこの
// メソッドを根拠にautofocusを手放し要約を描画するため、誤って答える状態があると失敗の理由が
// 通知されないまま残る。
func TestFormErrors_HasErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		errors *FormErrors
		want   bool
	}{
		{name: "nil", errors: nil, want: false},
		{name: "空", errors: NewFormErrors(model.NewValidationError()), want: false},
		{name: "グローバルエラーのみ", errors: &FormErrors{Global: []string{"エラー"}}, want: true},
		{
			name:   "フィールドエラーのみ",
			errors: &FormErrors{Fields: map[string][]string{"title": {"入力してください"}}},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.errors.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v、期待値 = %v", got, tt.want)
			}
		})
	}
}

// TestFormErrors_NilReceiverはnilの射影に対してもフィールドの参照が答えることを検証
// する。ページによってはデータ側のヘルパー (nilを1度だけ確認する) を経由し、別のページは
// マークアップから直接呼ぶため、どちらの読み方も安全である必要がある。
func TestFormErrors_NilReceiver(t *testing.T) {
	t.Parallel()

	var errors *FormErrors

	if errors.HasFieldError("title") {
		t.Error("HasFieldError() = true、期待値 = false")
	}
	if got := errors.GetFieldErrors("title"); got != nil {
		t.Errorf("GetFieldErrors() = %v、期待値 = nil", got)
	}
}
