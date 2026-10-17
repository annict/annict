package legaldoc

import (
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		source      string
		want        string
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:   "段落内の改行をbrにする",
			source: "1文目です。\n2文目です。",
			want:   "<p>1文目です。<br>\n2文目です。</p>\n",
		},
		{
			name:   "見出しに日本語を残したIDを付ける",
			source: "## 第1条 (適用)",
			want:   "<h2 id=\"第1条-適用\">第1条 (適用)</h2>\n",
		},
		{
			name:   "英語の見出しのIDは小文字にする",
			source: "## Article 1 (Application)",
			want:   "<h2 id=\"article-1-application\">Article 1 (Application)</h2>\n",
		},
		{
			name:   "記号だけの見出しにも空でないIDを付ける",
			source: "## ・・・",
			want:   "<h2 id=\"heading\">・・・</h2>\n",
		},
		{
			name:   "同じ見出しが続くときはIDに連番を付ける",
			source: "## 附則\n\n## 附則\n\n## 附則",
			want:   "<h2 id=\"附則\">附則</h2>\n<h2 id=\"附則-1\">附則</h2>\n<h2 id=\"附則-2\">附則</h2>\n",
		},
		{
			name:   "表を描画する",
			source: "| 利用目的 | 対応する情報 |\n| --- | --- |\n| 本サービスの提供 | メールアドレス |",
			wantContain: []string{
				"<table>",
				"<th>利用目的</th>",
				"<td>本サービスの提供</td>",
			},
		},
		{
			name:   "空行で区切った項目と内容の対を1つの定義リストにする",
			source: "販売事業者\n: 運営者の氏名\n\n販売価格\n: 月額290円",
			want:   "<dl>\n<dt>販売事業者</dt>\n<dd>運営者の氏名</dd>\n<dt>販売価格</dt>\n<dd>月額290円</dd>\n</dl>\n",
		},
		{
			name:   "定義リストの内容の改行をbrにする",
			source: "返品について\n: 1文目\n  2文目",
			want:   "<dl>\n<dt>返品について</dt>\n<dd>1文目<br>\n2文目</dd>\n</dl>\n",
		},
		{
			name:   "URLを自動でリンクにする",
			source: "詳細は https://annict.com/privacy を参照",
			want:   "<p>詳細は <a href=\"https://annict.com/privacy\">https://annict.com/privacy</a> を参照</p>\n",
		},
		{
			name:   "メールアドレスを自動でリンクにする",
			source: "- メールアドレス: hello@annict.com",
			want:   "<ul>\n<li>メールアドレス: <a href=\"mailto:hello@annict.com\">hello@annict.com</a></li>\n</ul>\n",
		},
		{
			name:   "リンク記法の中のURLを二重にリンクしない",
			source: "([https://annict.com/privacy](https://annict.com/privacy))",
			want:   "<p>(<a href=\"https://annict.com/privacy\">https://annict.com/privacy</a>)</p>\n",
		},
		{
			name:   "インラインの生HTMLを出力しない",
			source: "<a href=\"https://example.com/\">リンク</a>",
			wantAbsent: []string{
				"<a ",
			},
		},
		{
			name:   "ブロックの生HTMLを出力しない",
			source: "<script>alert(1)</script>",
			wantAbsent: []string{
				"<script",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := renderMarkdown([]byte(tt.source))
			if err != nil {
				t.Fatalf("renderMarkdown()のエラー = %v", err)
			}

			if tt.want != "" && got != tt.want {
				t.Errorf("描画結果 = %q、期待値 = %q", got, tt.want)
			}
			for _, s := range tt.wantContain {
				if !strings.Contains(got, s) {
					t.Errorf("描画結果 = %q に %q が含まれていない", got, s)
				}
			}
			for _, s := range tt.wantAbsent {
				if strings.Contains(got, s) {
					t.Errorf("描画結果 = %q に %q が含まれている", got, s)
				}
			}
		})
	}
}

func TestRenderMarkdown_HeadingIDsPerDocument(t *testing.T) {
	t.Parallel()

	// 見出しIDの重複は文書ごとに判定するため、別の文書の同じ見出しには連番を付けない
	for range 2 {
		got, err := renderMarkdown([]byte("## 附則"))
		if err != nil {
			t.Fatalf("renderMarkdown()のエラー = %v", err)
		}

		want := "<h2 id=\"附則\">附則</h2>\n"
		if got != want {
			t.Errorf("描画結果 = %q、期待値 = %q", got, want)
		}
	}
}
