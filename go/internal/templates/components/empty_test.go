package components

import (
	"context"
	"strings"
	"testing"
)

// TestEmpty verifies that the empty state names what is missing in a heading one level below
// the page heading, and that the description is rendered only when the caller supplies one.
// Basecoat lays the header out as a flex column with a gap, so an empty <p> would leave a
// blank band under the heading.
//
// [Ja] TestEmpty は、空表示が何が無いのかをページ見出しの 1 段下の見出しで述べること、説明は
// 呼び出し側が渡したときだけ描画されることを検証する。Basecoat は header を gap 付きの flex
// カラムとして並べるため、空の <p> があると見出しの下に余白の帯が残る。
func TestEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		title       string
		description string
		want        []string
		notWant     []string
	}{
		{
			name:        "説明なし",
			title:       "エピソードはありません",
			description: "",
			want: []string{
				`<section class="empty">`,
				"<header>",
				"<h2>エピソードはありません</h2>",
			},
			notWant: []string{"<p>"},
		},
		{
			name:        "説明あり",
			title:       "エピソードはありません",
			description: "エピソードを登録するとここに表示されます。",
			want: []string{
				"<header><h2>エピソードはありません</h2><p>エピソードを登録するとここに表示されます。</p></header>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			if err := Empty(tt.title, tt.description).Render(context.Background(), &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			for _, expected := range tt.want {
				if !strings.Contains(html, expected) {
					t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", expected, html)
				}
			}

			for _, unexpected := range tt.notWant {
				if strings.Contains(html, unexpected) {
					t.Errorf("含まれてはいけない文字列が含まれています: %q\nHTML: %s", unexpected, html)
				}
			}
		})
	}
}
