package db_person_revisions

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestIndex_EditLinkは、人物の履歴の見出しに、編集者にだけ編集画面へ戻るリンクを出すことを検証する。
// 掲載停止中の人物は編集できないため、編集者にも出さない。
func TestIndex_EditLink(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name        string
		isCommitter bool
		delisted    bool
		want        bool
	}{
		{name: "編集者", isCommitter: true, want: true},
		{name: "編集者以外", isCommitter: false, want: false},
		{name: "掲載停止中の人物の編集者", isCommitter: true, delisted: true, want: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), "ja")
			var buf strings.Builder
			if err := Index(IndexPageData{
				Person:      viewmodel.DBPersonRevisionPerson{ID: 7, Name: "人物", Delisted: tt.delisted},
				Pagination:  viewmodel.NewPagination(1, 0, 30, "/db/people/7/revisions"),
				IsCommitter: tt.isCommitter,
			}).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}
			html := buf.String()

			// 「一覧に戻る」と同じアイコンを、装飾としてテキストの左に付ける。
			withIcon := `href="/db/people/7/edit" class="btn rounded-full" data-variant="outline" data-size="sm"><svg data-icon="inline-start" aria-hidden="true" focusable="false" xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" viewBox="0 0 256 256"><path d="M232,200a8`
			for _, s := range []string{`href="/db/people/7/edit"`, "編集に戻る", withIcon} {
				if got := strings.Contains(html, s); got != tt.want {
					t.Errorf("%qを含むか = %v、期待値 = %v", s, got, tt.want)
				}
			}
		})
	}
}
