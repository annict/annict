package db_people

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

func renderNew(t *testing.T, data NewPageData) string {
	t.Helper()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := New(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	return buf.String()
}

// TestNew_FreshFormは、新規フォームが入力の形式を説明として関連付けたtextareaに
// フォーカスし、一覧へ戻るリンクと送信先を持つことを検証する。
func TestNew_FreshForm(t *testing.T) {
	t.Parallel()

	html := renderNew(t, NewPageData{CSRFToken: "test-csrf-token"})

	expectedContents := []string{
		"<h1",
		"人物登録",
		`action="/db/people"`,
		`method="POST"`,
		`value="test-csrf-token"`,
		`<label for="rows" class="label">`,
		`id="rows-description"`,
		"1行に1人ずつ、名前と読みをカンマ (,) で区切って入力してください。読みは省略できます。",
		"徳川家康,とくがわいえやす",
		`aria-describedby="rows-description"`,
		"autofocus",
		`href="/db/people"`,
		"登録する",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	if strings.Contains(html, `aria-invalid="true"`) {
		t.Error("新規フォームのtextareaが不正と印付けられています")
	}
}

// TestNew_WithErrorsは、却下された送信のメッセージを行ごとに表示し、textareaが形式の
// 説明に続けてそれらを指し、送信された行を書き戻すことを検証する。
func TestNew_WithErrors(t *testing.T) {
	t.Parallel()

	html := renderNew(t, NewPageData{
		CSRFToken: "test-csrf-token",
		FormErrors: &viewmodel.FormErrors{
			Fields: map[string][]string{"rows": {
				"2 行目: 名前を入力してください",
				"3 行目: この名前の人物は既に登録されています",
			}},
		},
		Rows: "徳川家康\n,よみだけ\n織田信長",
	})

	expectedContents := []string{
		`role="alert"`,
		"2 行目: 名前を入力してください",
		"3 行目: この名前の人物は既に登録されています",
		`aria-invalid="true"`,
		// 形式の説明はエラーより先に読まれ、エラー時にも失われない。
		`aria-describedby="rows-description rows-error-1 rows-error-2"`,
		`id="rows-error-1"`,
		`id="rows-error-2"`,
		"徳川家康\n,よみだけ\n織田信長",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに%qが含まれていません", expected)
		}
	}

	// エラー要約が読み込み時にフォーカスを受け取るため、textareaが同時にautofocusを
	// 主張してはいけない。
	if strings.Count(html, "autofocus") != 1 {
		t.Errorf("autofocusの数 = %d、期待値 = 1 (エラー要約のみ)", strings.Count(html, "autofocus"))
	}
}
