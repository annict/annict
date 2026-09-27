package viewmodel

import "testing"

// TestHelpPageURLsは各ヘルプURLがWikinoのAnnictスペースを指し、その名前が示すページID・
// トピック番号を含むことを検証する。サイドバーとDBのフォームはこれらの文字列をそのまま埋め込む
// ため、パスを誤ると読者を存在しないページへ黙って送ってしまう。
func TestHelpPageURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "作品の編集ガイドライン",
			got:  HelpWorkEditingURL(),
			want: "https://wikino.app/s/annict/pages/" + helpWorkEditingPageID,
		},
		{
			name: "エピソードの編集ガイドライン",
			got:  HelpEpisodeEditingURL(),
			want: "https://wikino.app/s/annict/pages/" + helpEpisodeEditingPageID,
		},
		{
			name: "エピソードの一括作成の行の形式",
			got:  HelpEpisodeBulkCreateURL(),
			want: "https://wikino.app/s/annict/pages/" + helpEpisodeBulkCreatePageID,
		},
		{
			name: "開発者向けドキュメントのトピック",
			got:  DeveloperHelpTopicURL(),
			want: "https://wikino.app/s/annict/topics/5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.got != tt.want {
				t.Errorf("URL = %q、期待値 = %q", tt.got, tt.want)
			}
		})
	}
}
