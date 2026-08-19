package viewmodel

import (
	"slices"
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// TestNewFormErrors_Nil verifies that a nil validation error projects onto a nil FormErrors.
// Handlers pass the projection straight into the page data, and every form template treats a
// nil value as "this submit was not rejected", so a non-nil empty projection would render the
// error summary of a form nobody submitted.
//
// [Ja] TestNewFormErrors_Nil は nil のバリデーションエラーが nil の FormErrors へ射影される
// ことを検証する。ハンドラーは射影をそのままページのデータへ渡し、各フォームのテンプレートは
// nil を「この送信は却下されていない」として扱うため、空の非 nil を返すと誰も送信していない
// フォームのエラー要約が描画されてしまう。
func TestNewFormErrors_Nil(t *testing.T) {
	t.Parallel()

	if got := NewFormErrors(nil); got != nil {
		t.Errorf("NewFormErrors(nil) = %v, want nil", got)
	}
}

// TestNewFormErrors_MirrorsValidationError verifies that the projection reports the same
// messages as the domain error it was built from. Templates moved from *model.ValidationError
// to this type without changing how they read it, so a projection that dropped or reordered
// messages would change what a rejected submit shows without any template changing.
//
// [Ja] TestNewFormErrors_MirrorsValidationError は、射影が元のドメインのエラーと同じメッセージ
// を返すことを検証する。テンプレートは読み方を変えないまま *model.ValidationError から本型へ
// 移ったため、メッセージを落としたり並べ替えたりする射影は、どのテンプレートも変わらないまま
// 却下された送信の表示を変えてしまう。
func TestNewFormErrors_MirrorsValidationError(t *testing.T) {
	t.Parallel()

	ve := model.NewValidationError()
	ve.AddGlobal("フォーム全体のエラー")
	ve.AddField("title", "入力してください")
	ve.AddField("title", "50 文字以内で入力してください")
	ve.AddField("number", "数値で入力してください")

	got := NewFormErrors(ve)

	if !slices.Equal(got.Global, ve.Global) {
		t.Errorf("Global = %v, want %v", got.Global, ve.Global)
	}

	for _, field := range []string{"title", "number", "absent"} {
		if want := ve.HasFieldError(field); got.HasFieldError(field) != want {
			t.Errorf("HasFieldError(%q) = %v, want %v", field, got.HasFieldError(field), want)
		}
		if want := ve.GetFieldErrors(field); !slices.Equal(got.GetFieldErrors(field), want) {
			t.Errorf("GetFieldErrors(%q) = %v, want %v", field, got.GetFieldErrors(field), want)
		}
	}

	if !got.HasErrors() {
		t.Error("HasErrors() = false, want true")
	}
}

// TestFormErrors_HasErrors covers the three states a form template distinguishes: no
// projection at all, a projection holding only global messages, and one holding only field
// messages. The pages give up autofocus and render the summary on the strength of this method,
// so a state it answered wrongly would leave the reason for the failure unannounced.
//
// [Ja] TestFormErrors_HasErrors はフォームのテンプレートが区別する 3 つの状態 (射影が無い、
// グローバルメッセージだけを持つ、フィールドのメッセージだけを持つ) を扱う。各ページはこの
// メソッドを根拠に autofocus を手放し要約を描画するため、誤って答える状態があると失敗の理由が
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
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFormErrors_NilReceiver verifies that the field lookups answer for a nil projection.
// Some pages ask a helper on the page data (which checks for nil once) and others call these
// straight from the markup, so both readings have to be safe.
//
// [Ja] TestFormErrors_NilReceiver は nil の射影に対してもフィールドの参照が答えることを検証
// する。ページによってはデータ側のヘルパー (nil を 1 度だけ確認する) を経由し、別のページは
// マークアップから直接呼ぶため、どちらの読み方も安全である必要がある。
func TestFormErrors_NilReceiver(t *testing.T) {
	t.Parallel()

	var errors *FormErrors

	if errors.HasFieldError("title") {
		t.Error("HasFieldError() = true, want false")
	}
	if got := errors.GetFieldErrors("title"); got != nil {
		t.Errorf("GetFieldErrors() = %v, want nil", got)
	}
}
