package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestDeletionNew verifies the confirmation page renders the work title as its heading and in
// the confirm message, wraps the confirmation in the shared content card, and submits a CSRF
// token to the work endpoint as a DELETE through the method override.
//
// [Ja] TestDeletionNew は確認ページが作品タイトルを見出しと確認メッセージに描画し、確認内容を
// 共有のコンテンツカードに載せ、メソッドオーバーライドで DELETE として CSRF トークン付きで
// 作品のエンドポイントへ送信することを検証する。
func TestDeletionNew(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := DeletionNewPageData{
		CSRFToken: "test-csrf",
		WorkID:    viewmodel.WorkID(42),
		Title:     "確認対象アニメ",
		ReturnTo:  "/db/search?q=test",
	}

	var buf strings.Builder
	if err := DeletionNew(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	expectedContents := []string{
		// The form posts to the work itself, and the method override turns the POST into
		// the DELETE the work endpoint serves.
		//
		// [Ja] フォームは作品自身へ POST し、メソッドオーバーライドがそれを作品エンドポイントの
		// DELETE に変換する。
		`action="/db/works/42"`,
		`method="POST"`,
		`name="_method" value="DELETE"`,
		`name="csrf_token"`,
		`value="test-csrf"`,
		// The work title is the page heading and is also interpolated into the
		// confirmation message.
		//
		// [Ja] 作品タイトルがページ見出しになり、確認メッセージにも埋め込まれる。
		">確認対象アニメ</h1>",
		"「確認対象アニメ」を削除しますか？",
		// The confirmation sits in the same card container as the other /db pages.
		//
		// [Ja] 確認内容は他の /db 画面と同じカードコンテナに載る。
		`class="card`,
		// The execute button carries the same destructive variant as the delete button in
		// the work list, and the cancel link the outline variant.
		//
		// [Ja] 実行ボタンは作品一覧の削除ボタンと同じ destructive、キャンセルリンクは outline
		// のバリアントを持つ。
		`class="btn rounded-full" data-variant="destructive"`,
		`class="btn rounded-full" data-variant="outline"`,
		// The cancel link and the form both carry the listing the reader came from, so
		// leaving the confirmation and completing it land on the same page.
		//
		// [Ja] キャンセルリンクとフォームの双方が読み手の来た一覧を持ち回るため、確認を
		// やめた場合も完了した場合も同じページに着地する。
		`<a href="/db/search?q=test"`,
		`name="return_to" value="/db/search?q=test"`,
		// The header renders the sidebar toggle, wired to the sidebar at every viewport size.
		//
		// [Ja] ヘッダーはサイドバートグルを描画する。サイドバーに結線され、
		// 全画面幅で利用できる。
		`data-sidebar-toggle="db-sidebar"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("response doesn't contain expected string: %q", expected)
		}
	}
}

// TestDeletionNew_BlankTitleFallsBackToPageTitle verifies the heading and the confirmation
// message both fall back to the page title when the work title is blank, so the page never
// renders an empty <h1> and never asks about an unnamed target.
//
// [Ja] TestDeletionNew_BlankTitleFallsBackToPageTitle は作品タイトルが空のときに見出しと
// 確認文がどちらもページタイトルへフォールバックし、空の <h1> を描画せず、対象を名指し
// できない確認文にもならないことを検証する。
func TestDeletionNew_BlankTitleFallsBackToPageTitle(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := DeletionNewPageData{
		CSRFToken: "test-csrf",
		WorkID:    viewmodel.WorkID(42),
		Title:     "",
	}

	var buf strings.Builder
	if err := DeletionNew(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, ">作品削除</h1>") {
		t.Error("response doesn't fall back to the page title in the heading")
	}

	if !strings.Contains(html, "「作品削除」を削除しますか？") {
		t.Error("response doesn't fall back to the page title in the confirmation message")
	}
}
