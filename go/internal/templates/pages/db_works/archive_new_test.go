package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestArchiveNew verifies the confirmation page renders the work title into the confirm
// message and posts to the work's archive endpoint with a CSRF token.
//
// [Ja] TestArchiveNew は確認ページが作品タイトルを確認メッセージに埋め込み、CSRF トークン付き
// で作品の非公開エンドポイントへ POST することを検証する。
func TestArchiveNew(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := ArchiveNewPageData{
		CSRFToken: "test-csrf",
		WorkID:    viewmodel.WorkID(42),
		Title:     "確認対象アニメ",
	}

	var buf strings.Builder
	if err := ArchiveNew(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	expectedContents := []string{
		`action="/db/works/42/archive"`,
		`method="POST"`,
		`name="csrf_token"`,
		`value="test-csrf"`,
		// The title is interpolated into the confirmation message.
		//
		// [Ja] タイトルが確認メッセージに埋め込まれる。
		"確認対象アニメ",
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}
}
