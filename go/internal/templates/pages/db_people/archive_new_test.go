package db_people

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestArchiveNewはアーカイブ確認ページが人物の名前を見出しと確認メッセージに描画し、CSRFトークンと
// 読み手の来た一覧を持ち回ってアーカイブのエンドポイントへ送信することを検証する。
func TestArchiveNew(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	var buf strings.Builder
	if err := ArchiveNew(ArchiveNewPageData{
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
		`name="csrf_token" value="test-csrf"`,
		">確認対象の人物</h1>",
		"「確認対象の人物」を非公開にしますか？",
		"非公開にすると、この人物は人物ページや作品のキャスト・スタッフに表示されなくなります。",
		`class="btn rounded-full" data-variant="warning"`,
		`<a href="/db/search?q=test"`,
		`name="return_to" value="/db/search?q=test"`,
		`data-sidebar-toggle="db-sidebar"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}

	// アーカイブはPOSTのまま送るため、メソッドオーバーライドを持たない。
	if strings.Contains(html, `name="_method"`) {
		t.Error("アーカイブのフォームがメソッドオーバーライドを持っています")
	}
}
