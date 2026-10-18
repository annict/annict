package db_people

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestUndelistingNewは掲載停止解除確認ページが人物の名前と対応記録を描画し、メソッドオーバーライドで
// DELETEを運んで掲載停止のエンドポイントへ送信することを検証する。
func TestUndelistingNew(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	var buf strings.Builder
	if err := UndelistingNew(UndelistingNewPageData{
		CSRFToken: "test-csrf",
		Person:    viewmodel.DBPersonDelistingTarget{ID: 42, Name: "確認対象の人物", DelistMessage: "本人から申出\n対応日: 2026-10-18"},
		ReturnTo:  "/db/people",
	}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	for _, expected := range []string{
		`action="/db/people/42/delisting"`,
		`name="_method" value="DELETE"`,
		`name="csrf_token" value="test-csrf"`,
		">確認対象の人物</h1>",
		"「確認対象の人物」の掲載停止を解除しますか？",
		`aria-labelledby="delist_message-heading"`,
		"本人から申出\n対応日: 2026-10-18",
		`class="btn rounded-full" data-variant="success"`,
		`name="return_to" value="/db/people"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}
}

// TestUndelistingNew_NoDelistMessageは、対応記録が無い人物では、記録が無いことを示す文言を描画する
// ことを検証する。
func TestUndelistingNew_NoDelistMessage(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	var buf strings.Builder
	if err := UndelistingNew(UndelistingNewPageData{
		CSRFToken: "test-csrf",
		Person:    viewmodel.DBPersonDelistingTarget{ID: 42, Name: "確認対象の人物"},
		ReturnTo:  "/db/people",
	}).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	if expected := "対応記録はありません"; !strings.Contains(buf.String(), expected) {
		t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
	}
}
