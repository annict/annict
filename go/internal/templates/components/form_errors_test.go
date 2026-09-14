package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// summaryFieldsは要約を有効にするために呼び出し側が渡すフィールド一覧。要約はこの順序に
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
			// 呼び出し側が挙げていないフィールドのエラーはラベルもアンカーも無く、要約が
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

// TestFormErrors_Summaryは、送信の失敗がフォーム冒頭に要約されることを検証する。各
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

	// 項目は地の文ではなくリンクの一覧であるため、行高24pxで各リンク自身のボックスを
	// タップ領域の最小値にし、paddingによる文字のずれを避ける。
	if !strings.Contains(html, `class="inline-block align-top leading-6 underline"`) {
		t.Errorf("要約のリンクはタップ領域を確保する必要があります\nHTML: %s", html)
	}

	// inline-blockの項目が折り返すと、alertの既定である内側のマーカーだけが1行上に
	// 取り残されることがある。
	if !strings.Contains(html, `<ul class="list-outside space-y-1 ps-5">`) {
		t.Errorf("要約のマーカーは内容ボックスの外に置く必要があります\nHTML: %s", html)
	}

	// ページは要約から始まるため、落ちたフィールドがフォームの下の方にあっても、送信が
	// 失敗した理由へ最初に到達できる。
	for _, want := range []string{`tabindex="-1"`, "autofocus", `role="alert"`} {
		if !strings.Contains(html, want) {
			t.Errorf("要約はフォーカスを受け取る必要があります: %q\nHTML: %s", want, html)
		}
	}

	// FormErrorsはフィールドエラーをmapで保持するため、順序は走査ではなく
	// 呼び出し側のフィールド一覧から決まる必要がある。
	if strings.Index(html, "#title") > strings.Index(html, "#sc_tid") {
		t.Errorf("要約はフォームの表示順に並べる必要があります\nHTML: %s", html)
	}
}

// TestFormErrors_SummaryWithGlobalは、フォーム全体に紐づくメッセージがリンク無しで要約
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

// TestFormErrors_SummaryLocalesは、見出しとラベル / メッセージの組み合わせの両方が翻訳
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

			// 項目はどのロケールで描画してもフィールドのラベルとメッセージを対にする。
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
