package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestUnarchiveNew verifies the confirmation page renders the work title as its heading and in
// the confirm message, wraps the confirmation in the shared content card, and submits a CSRF
// token to the work's archive endpoint as a DELETE through the method override.
//
// [Ja] TestUnarchiveNew は確認ページが作品タイトルを見出しと確認メッセージに描画し、確認内容を
// 共有のコンテンツカードに載せ、メソッドオーバーライドで DELETE として CSRF トークン付きで
// 作品の非公開エンドポイントへ送信することを検証する。
func TestUnarchiveNew(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := UnarchiveNewPageData{
		CSRFToken: "test-csrf",
		WorkID:    viewmodel.WorkID(42),
		Title:     "確認対象アニメ",
		ReturnTo:  "/db/search?q=test",
	}

	var buf strings.Builder
	if err := UnarchiveNew(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	expectedContents := []string{
		// Publishing a work is deleting its archive, so the form posts to the archive
		// endpoint and the method override turns the POST into the DELETE it serves.
		//
		// [Ja] 作品の公開はそのアーカイブの削除であるため、フォームは非公開エンドポイントへ
		// POST し、メソッドオーバーライドがそれを同エンドポイントの DELETE に変換する。
		`action="/db/works/42/archive"`,
		`method="POST"`,
		`name="_method" value="DELETE"`,
		`name="csrf_token"`,
		`value="test-csrf"`,
		// The work title is the page heading and is also interpolated into the
		// confirmation message.
		//
		// [Ja] 作品タイトルがページ見出しになり、確認メッセージにも埋め込まれる。
		">確認対象アニメ</h1>",
		"「確認対象アニメ」を公開しますか？",
		// The confirmation sits in the same card container as the other /db pages.
		//
		// [Ja] 確認内容は他の /db 画面と同じカードコンテナに載る。
		`class="card`,
		// The execute button carries the same success variant as the publish button in the
		// work list, and the cancel link the outline variant.
		//
		// [Ja] 実行ボタンは作品一覧の公開ボタンと同じ success、キャンセルリンクは outline の
		// バリアントを持つ。
		`class="btn rounded-full" data-variant="success"`,
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

// TestUnarchiveNew_BlankTitleFallsBackToPageTitle verifies the heading and the confirmation
// message both fall back to the page title when the work title is blank, so the page never
// renders an empty <h1> and never asks about an unnamed target.
//
// [Ja] TestUnarchiveNew_BlankTitleFallsBackToPageTitle は作品タイトルが空のときに見出しと
// 確認文がどちらもページタイトルへフォールバックし、空の <h1> を描画せず、対象を名指し
// できない確認文にもならないことを検証する。
func TestUnarchiveNew_BlankTitleFallsBackToPageTitle(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := UnarchiveNewPageData{
		CSRFToken: "test-csrf",
		WorkID:    viewmodel.WorkID(42),
		Title:     "",
	}

	var buf strings.Builder
	if err := UnarchiveNew(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, ">作品公開</h1>") {
		t.Error("response doesn't fall back to the page title in the heading")
	}

	if !strings.Contains(html, "「作品公開」を公開しますか？") {
		t.Error("response doesn't fall back to the page title in the confirmation message")
	}
}
