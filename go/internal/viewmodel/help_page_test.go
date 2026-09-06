package viewmodel

import "testing"

// TestHelpPageURLs verifies every help URL points at the Annict space on Wikino and carries
// the id or number of the page or topic it names. The sidebar and the DB forms embed these
// strings verbatim, so a wrong path silently sends readers to a page that does not exist.
//
// [Ja] TestHelpPageURLsは各ヘルプURLがWikinoのAnnictスペースを指し、その名前が示すページID・
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
				t.Errorf("URL = %q, want %q", tt.got, tt.want)
			}
		})
	}
}
