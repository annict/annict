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

// TestPopular_Castsは人気作品ページのキャスト欄がキャラクター名を並べ、
// キャラクター名が空のキャストはキャスト名で表示することを検証する。
func TestPopular_Casts(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	err := Popular(context.Background(), []viewmodel.Work{{
		ID:    1,
		Title: "Test Work",
		Casts: []viewmodel.Cast{
			{ID: 1, Name: "キャスト名1", CharacterName: "キャラクター1", PersonName: "声優1"},
			{ID: 2, Name: "キャスト名2", PersonName: "声優2"},
		},
	}}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Popular().Render()のエラー = %v", err)
	}

	html := buf.String()
	if expected := "キャラクター1, キャスト名2"; !strings.Contains(html, expected) {
		t.Errorf("HTMLにキャスト欄の表示 %q が含まれていません", expected)
	}
}
