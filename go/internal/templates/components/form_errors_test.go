package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// summaryFields is the field list a caller passes to turn the summary on. Its order is what
// the summary has to follow.
//
// [Ja] summaryFields は要約を有効にするために呼び出し側が渡すフィールド一覧。要約はこの順序に
// 従う必要がある。
var summaryFields = []FormErrorField{
	{Name: "title", Label: "タイトル"},
	{Name: "media", Label: "メディア"},
	{Name: "sc_tid", Label: "しょぼいカレンダー"},
}

func TestFormErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		data          FormErrorsData
		wantContains  []string
		wantNotRender bool
	}{
		{
			name: "グローバルエラーが1つ",
			data: FormErrorsData{
				Errors: &viewmodel.FormErrors{
					Global: []string{"ログインに失敗しました"},
				},
			},
			wantContains: []string{
				`<div class="alert" data-variant="destructive">`,
				`<h2>ログインに失敗しました</h2>`,
			},
		},
		{
			name: "グローバルエラーが複数",
			data: FormErrorsData{
				Errors: &viewmodel.FormErrors{
					Global: []string{
						"エラー1",
						"エラー2",
						"エラー3",
					},
				},
			},
			wantContains: []string{
				`<h2>エラー1</h2>`,
				`<h2>エラー2</h2>`,
				`<h2>エラー3</h2>`,
			},
		},
		{
			name: "フィールドを渡さなければフィールドエラーは表示しない",
			data: FormErrorsData{
				Errors: &viewmodel.FormErrors{
					Fields: map[string][]string{
						"email": {"メールアドレスが不正です"},
					},
				},
			},
			wantNotRender: true,
		},
		{
			name:          "formErrorsがnilの場合は何も表示しない",
			data:          FormErrorsData{},
			wantNotRender: true,
		},
		{
			name: "formErrorsが空の場合は何も表示しない",
			data: FormErrorsData{
				Errors: &viewmodel.FormErrors{
					Global: []string{},
				},
			},
			wantNotRender: true,
		},
		{
			name: "フィールドを渡してもエラーが無ければ何も表示しない",
			data: FormErrorsData{
				Errors: &viewmodel.FormErrors{},
				Fields: summaryFields,
			},
			wantNotRender: true,
		},
		{
			// A field error the caller did not list has neither a label nor an anchor, so the
			// summary cannot name it. Falling back to the global-only rendering keeps the
			// component from emitting an empty summary.
			//
			// [Ja] 呼び出し側が挙げていないフィールドのエラーはラベルもアンカーも無く、要約が
			// 名指しできない。グローバルのみの描画に落ちることで、空の要約を出さずに済む。
			name: "一覧に無いフィールドのエラーだけなら要約を出さない",
			data: FormErrorsData{
				Errors: &viewmodel.FormErrors{
					Fields: map[string][]string{
						"season_year": {"整数で入力してください"},
					},
				},
				Fields: summaryFields,
			},
			wantNotRender: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")

			var buf strings.Builder
			if err := FormErrors(tt.data).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			if tt.wantNotRender {
				if strings.TrimSpace(html) != "" {
					t.Errorf("何も表示されないはずだが、HTMLが生成されました: %s", html)
				}
				return
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(html, want) {
					t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", want, html)
				}
			}
		})
	}
}

// TestFormErrors_Summary verifies that a failed submit is summarised at the top of the form:
// every error is named with the label the form shows and linked to the input it belongs to,
// so the failing field can be reached without hunting for it.
//
// [Ja] TestFormErrors_Summary は、送信の失敗がフォーム冒頭に要約されることを検証する。各
// エラーはフォームが表示するラベルで名指しされ、対象の入力欄へリンクされるため、落ちた
// フィールドを探し回らずに到達できる。
func TestFormErrors_Summary(t *testing.T) {
	t.Parallel()

	formErrors := &viewmodel.FormErrors{
		Fields: map[string][]string{
			"sc_tid": {"整数で入力してください"},
			"title":  {"入力してください"},
		},
	}

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := FormErrors(FormErrorsData{Errors: formErrors, Fields: summaryFields}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	for _, want := range []string{
		"入力内容にエラーがあります",
		`href="#title"`,
		"タイトル: 入力してください",
		`href="#sc_tid"`,
		"しょぼいカレンダー: 整数で入力してください",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", want, html)
		}
	}

	// The entries are a list of links rather than prose, so a 24px line height makes each
	// link's own box meet the minimum tap target without padding shifting its text.
	//
	// [Ja] 項目は地の文ではなくリンクの一覧であるため、行高 24px で各リンク自身のボックスを
	// タップ領域の最小値にし、padding による文字のずれを避ける。
	if !strings.Contains(html, `class="inline-block align-top leading-6 underline"`) {
		t.Errorf("要約のリンクはタップ領域を確保する必要があります\nHTML: %s", html)
	}

	// An inline-block entry can strand the alert's default inside marker on the line above
	// when the entry wraps.
	//
	// [Ja] inline-block の項目が折り返すと、alert の既定である内側のマーカーだけが 1 行上に
	// 取り残されることがある。
	if !strings.Contains(html, `<ul class="list-outside space-y-1 ps-5">`) {
		t.Errorf("要約のマーカーは内容ボックスの外に置く必要があります\nHTML: %s", html)
	}

	// The summary is where the page starts out, so the reason the submit failed is reached
	// first even when the failing field is far down the form.
	//
	// [Ja] ページは要約から始まるため、落ちたフィールドがフォームの下の方にあっても、送信が
	// 失敗した理由へ最初に到達できる。
	for _, want := range []string{`tabindex="-1"`, "autofocus", `role="alert"`} {
		if !strings.Contains(html, want) {
			t.Errorf("要約はフォーカスを受け取る必要があります: %q\nHTML: %s", want, html)
		}
	}

	// FormErrors keeps its field errors in a map, so the order has to come from the
	// caller's field list rather than from iteration.
	//
	// [Ja] FormErrors はフィールドエラーを map で保持するため、順序は走査ではなく
	// 呼び出し側のフィールド一覧から決まる必要がある。
	if strings.Index(html, "#title") > strings.Index(html, "#sc_tid") {
		t.Errorf("要約はフォームの表示順に並べる必要があります\nHTML: %s", html)
	}
}

// TestFormErrors_SummaryWithGlobal verifies that a message belonging to the form as a whole
// joins the summary without a link: it has no single field to point at.
//
// [Ja] TestFormErrors_SummaryWithGlobal は、フォーム全体に紐づくメッセージがリンク無しで要約
// に並ぶことを検証する。指し示す先の単一のフィールドが無いため。
func TestFormErrors_SummaryWithGlobal(t *testing.T) {
	t.Parallel()

	formErrors := &viewmodel.FormErrors{
		Global: []string{"保存に失敗しました"},
		Fields: map[string][]string{"title": {"入力してください"}},
	}

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := FormErrors(FormErrorsData{Errors: formErrors, Fields: summaryFields}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	for _, want := range []string{"保存に失敗しました", "タイトル: 入力してください"} {
		if !strings.Contains(html, want) {
			t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", want, html)
		}
	}

	if strings.Contains(html, `<h2>保存に失敗しました</h2>`) {
		t.Errorf("要約があるときはグローバルエラーも要約に含める必要があります\nHTML: %s", html)
	}
}

// TestFormErrors_SummaryLocales verifies that the heading and the label / message pairing are
// both translated, so the summary reads in the language the rest of the form uses.
//
// [Ja] TestFormErrors_SummaryLocales は、見出しとラベル / メッセージの組み合わせの両方が翻訳
// されることを検証する。要約がフォームの他の部分と同じ言語で読めるようにするため。
func TestFormErrors_SummaryLocales(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{
			name:   "日本語",
			locale: "ja",
			want:   "入力内容にエラーがあります",
		},
		{
			name:   "英語",
			locale: "en",
			want:   "There is a problem with your input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formErrors := &viewmodel.FormErrors{
				Fields: map[string][]string{"title": {"入力してください"}},
			}

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			var buf strings.Builder
			if err := FormErrors(FormErrorsData{Errors: formErrors, Fields: summaryFields}).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			if !strings.Contains(html, tt.want) {
				t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", tt.want, html)
			}

			// The entry pairs the field's label with its message, whichever locale renders it.
			//
			// [Ja] 項目はどのロケールで描画してもフィールドのラベルとメッセージを対にする。
			if !strings.Contains(html, "タイトル: 入力してください") {
				t.Errorf("要約の項目はラベルとメッセージを対にする必要があります\nHTML: %s", html)
			}
		})
	}
}

func TestFormErrors_HTMLStructure(t *testing.T) {
	t.Parallel()

	data := FormErrorsData{
		Errors: &viewmodel.FormErrors{
			Global: []string{"エラーメッセージ"},
		},
	}

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := FormErrors(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// 正しいHTML構造を持っているか確認
	expectedStructure := []string{
		`<div class="alert" data-variant="destructive">`,
		`<h2>`,
		`</h2>`,
		`</div>`,
	}

	for _, expected := range expectedStructure {
		if !strings.Contains(html, expected) {
			t.Errorf("期待するHTML構造が含まれていません: %q\nHTML: %s", expected, html)
		}
	}
}
