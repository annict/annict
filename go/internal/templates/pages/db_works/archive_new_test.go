package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestArchiveNewは確認ページが作品タイトルを見出しと確認メッセージに描画し、確認内容を
// 共有のコンテンツカードに載せ、CSRFトークン付きで作品の非公開エンドポイントへPOSTする
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
		// 作品タイトルがページ見出しになり、確認メッセージにも埋め込まれる。
		">確認対象アニメ</h1>",
		// 確認メッセージの後ろに、非公開にすると何が起きるか、どう戻せるかの説明が続く。
		// エピソードの非公開確認と揃える。
		"非公開にすると、この作品は公開ページに表示されなくなります。あとから一覧の「公開」から元に戻せます。",
		// 確認内容は他の /db画面と同じカードコンテナに載る。
		`class="card`,
		// 実行ボタンは作品一覧の非公開リンクと同じwarning、キャンセルリンクはoutline
		// のバリアントを持つ。
		`class="btn rounded-full" data-variant="warning"`,
		`class="btn rounded-full" data-variant="outline"`,
		// キャンセルリンクとフォームの双方が読み手の来た一覧を持ち回るため、確認を
		// やめた場合も完了した場合も同じページに着地する。
		`<a href="/db/search?q=test"`,
		`name="return_to" value="/db/search?q=test"`,
		// ヘッダーはサイドバートグルを描画する。サイドバーに結線され、
		// 全画面幅で利用できる。
		`data-sidebar-toggle="db-sidebar"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("レスポンスに含まれていない文字列 = %q", expected)
		}
	}

	// タイトル行はキャンセルリンクと同じ行き先を重複して持たない。
	if strings.Contains(html, "一覧に戻る") {
		t.Error("削除したはずの一覧へ戻るリンクがレスポンスに残っている")
	}
}

// TestArchiveNew_BlankTitleFallsBackToPageTitleは作品タイトルが空のときに見出しと
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
		t.Error("見出しがページタイトルにフォールバックしていない")
	}

	if !strings.Contains(html, "「作品非公開」を非公開にしますか？") {
		t.Error("確認メッセージがページタイトルにフォールバックしていない")
	}
}
