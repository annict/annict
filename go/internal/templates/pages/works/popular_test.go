package works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/viewmodel"
)

// TestPopular_WorkImageは公開の人気作品ページが共通の3:4作品画像枠を使い、
// 元画像を切り抜かずに表示することを検証する。
func TestPopular_WorkImage(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	err := Popular(context.Background(), []viewmodel.Work{{
		ID:       1,
		Title:    "Test Work",
		ImageURL: "https://example.com/work.jpg",
	}}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Popular().Render()のエラー = %v", err)
	}

	html := buf.String()
	for _, expected := range []string{
		`src="https://example.com/work.jpg"`,
		`width="280"`,
		`height="373"`,
		`class="w-full h-full object-contain"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("HTMLに必要な作品画像属性が含まれていません: %q", expected)
		}
	}
	if strings.Contains(html, "object-cover") {
		t.Error("作品画像は切り抜くobject-coverを使用してはいけません")
	}
}
