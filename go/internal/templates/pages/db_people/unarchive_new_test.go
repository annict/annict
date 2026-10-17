package db_people

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestUnarchiveNewはアーカイブ解除確認ページが人物の名前を見出しと確認メッセージに描画し、
// CSRFトークンと読み手の来た一覧を持ち回ってアーカイブ解除のエンドポイントへ送信することを検証する。
func TestUnarchiveNew(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	var buf strings.Builder
	if err := UnarchiveNew(UnarchiveNewPageData{
		CSRFToken: "test-csrf",
		PersonID:  viewmodel.PersonID(42),
		Name:      "確認対象の人物",
		ReturnTo:  "/db/search?q=test",
	}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	for _, expected := range []string{
		`action="/db/people/42/archive"`,
		`method="POST"`,
		`name="_method" value="DELETE"`,
		`name="csrf_token" value="test-csrf"`,
		">確認対象の人物</h1>",
		"「確認対象の人物」のアーカイブを解除しますか？",
		`class="btn rounded-full" data-variant="success"`,
		`<a href="/db/search?q=test"`,
		`name="return_to" value="/db/search?q=test"`,
		`data-sidebar-toggle="db-sidebar"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}
