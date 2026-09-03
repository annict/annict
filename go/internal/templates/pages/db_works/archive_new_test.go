package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestArchiveNew verifies the confirmation page renders the work title as its heading and in
// the confirm message, wraps the confirmation in the shared content card, and posts to the
// work's archive endpoint with a CSRF token.
//
// [Ja] TestArchiveNew は確認ページが作品タイトルを見出しと確認メッセージに描画し、確認内容を
// 共有のコンテンツカードに載せ、CSRF トークン付きで作品の非公開エンドポイントへ POST する
// ことを検証する。
func TestArchiveNew(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := ArchiveNewPageData{
		CSRFToken: "test-csrf",
		WorkID:    viewmodel.WorkID(42),
		Title:     "確認対象アニメ",
		ReturnTo:  "/db/search?q=test",
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
		// The work title is the page heading and is also interpolated into the
		// confirmation message.
		//
		// [Ja] 作品タイトルがページ見出しになり、確認メッセージにも埋め込まれる。
		">確認対象アニメ</h1>",
		// The confirmation is followed by the explanation of what archiving does and how to
		// undo it, matching the episode archive confirmation.
		//
		// [Ja] 確認メッセージの後ろに、非公開にすると何が起きるか、どう戻せるかの説明が続く。
		// エピソードの非公開確認と揃える。
		"非公開にすると、この作品は公開ページに表示されなくなります。あとから一覧の「公開」から元に戻せます。",
		// The confirmation sits in the same card container as the other /db pages.
		//
		// [Ja] 確認内容は他の /db 画面と同じカードコンテナに載る。
		`class="card`,
		// The execute button carries the same warning variant as the archive link in the
		// work list, and the cancel link the outline variant.
		//
		// [Ja] 実行ボタンは作品一覧の非公開リンクと同じ warning、キャンセルリンクは outline
		// のバリアントを持つ。
		`class="btn rounded-full" data-variant="warning"`,
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

	// The title row no longer duplicates the cancel link's destination.
	//
	// [Ja] タイトル行はキャンセルリンクと同じ行き先を重複して持たない。
	if strings.Contains(html, "一覧に戻る") {
		t.Error("response still contains the removed back-to-list link")
	}
}

// TestArchiveNew_BlankTitleFallsBackToPageTitle verifies the heading and the confirmation
// message both fall back to the page title when the work title is blank, so the page never
// renders an empty <h1> and never asks about an unnamed target.
//
// [Ja] TestArchiveNew_BlankTitleFallsBackToPageTitle は作品タイトルが空のときに見出しと
// 確認文がどちらもページタイトルへフォールバックし、空の <h1> を描画せず、対象を名指し
// できない確認文にもならないことを検証する。
func TestArchiveNew_BlankTitleFallsBackToPageTitle(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")
	data := ArchiveNewPageData{
		CSRFToken: "test-csrf",
		WorkID:    viewmodel.WorkID(42),
		Title:     "",
	}

	var buf strings.Builder
	if err := ArchiveNew(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, ">作品非公開</h1>") {
		t.Error("response doesn't fall back to the page title in the heading")
	}

	if !strings.Contains(html, "「作品非公開」を非公開にしますか？") {
		t.Error("response doesn't fall back to the page title in the confirmation message")
	}
}
