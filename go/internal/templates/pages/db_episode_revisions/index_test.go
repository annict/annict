package db_episode_revisions

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestIndex_EditLinkは、エピソードの履歴の見出しに、編集者にだけエピソードの編集画面へ戻るリンクを出すことを検証する。
func TestIndex_EditLink(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name        string
		locale      string
		isCommitter bool
		want        string
	}{
		{name: "編集者", locale: "ja", isCommitter: true, want: "編集に戻る"},
		{name: "英語の編集者", locale: "en", isCommitter: true, want: "Back to Edit"},
		{name: "編集者以外", locale: "ja", isCommitter: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			var buf strings.Builder
			if err := Index(IndexPageData{
				Episode: viewmodel.DBEpisodeRevisionEpisode{
					ID:    9,
					Label: "第1話",
					Work:  viewmodel.DBWorkRevisionWork{ID: 3, Title: "作品"},
				},
				Pagination:  viewmodel.NewPagination(1, 0, 30, "/db/episodes/9/revisions"),
				IsCommitter: tt.isCommitter,
			}).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}
			html := buf.String()

			link := `href="/db/episodes/9/edit"`
			if got := strings.Contains(html, link); got != (tt.want != "") {
				t.Errorf("%qを含むか = %v、期待値 = %v", link, got, tt.want != "")
			}
			if tt.want != "" && !strings.Contains(html, tt.want) {
				t.Errorf("リンクの文言%qが含まれていない", tt.want)
			}
			// 「一覧に戻る」と同じアイコンを、装飾としてテキストの左に付ける。
			withIcon := link + ` class="btn rounded-full" data-variant="outline" data-size="sm"><svg data-icon="inline-start" aria-hidden="true" focusable="false" xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" viewBox="0 0 256 256"><path d="M232,200a8`
			if got := strings.Contains(html, withIcon); got != (tt.want != "") {
				t.Errorf("アイコン付きのリンクを含むか = %v、期待値 = %v", got, tt.want != "")
			}
		})
	}
}
