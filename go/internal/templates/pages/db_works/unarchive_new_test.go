package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestUnarchiveNewは確認ページが作品タイトルを見出しと確認メッセージに描画し、確認内容を
// 共有のコンテンツカードに載せ、メソッドオーバーライドでDELETEとしてCSRFトークン付きで
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
		// 作品の公開はそのアーカイブの削除であるため、フォームは非公開エンドポイントへ
		// POSTし、メソッドオーバーライドがそれを同エンドポイントのDELETEに変換する。
		`action="/db/works/42/archive"`,
		`method="POST"`,
		`name="_method" value="DELETE"`,
		`name="csrf_token"`,
		`value="test-csrf"`,
		// 作品タイトルがページ見出しになり、確認メッセージにも埋め込まれる。
		">確認対象アニメ</h1>",
		"「確認対象アニメ」を公開しますか？",
		// 確認内容は他の /db画面と同じカードコンテナに載る。
		`class="card`,
		// 実行ボタンは作品一覧の公開ボタンと同じsuccess、キャンセルリンクはoutlineの
		// バリアントを持つ。
		`class="btn rounded-full" data-variant="success"`,
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
}

// TestUnarchiveNew_BlankTitleFallsBackToPageTitleは作品タイトルが空のときに見出しと
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
		t.Error("見出しがページタイトルにフォールバックしていない")
	}

	if !strings.Contains(html, "「作品公開」を公開しますか？") {
		t.Error("確認メッセージがページタイトルにフォールバックしていない")
	}
}
