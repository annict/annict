package db_works

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/viewmodel"
)

// TestNew_LabelExternalLinks verifies that an external-link icon is rendered next to a
// label whose field has a value, and that the synopsis textareas use 10 rows like Rails.
//
// [Ja] TestNew_LabelExternalLinks は、値が入っているラベルの横に外部リンクアイコンが描画され、
// synopsis テキストエリアが Rails と同じ行数 (10) になることをテストする。
func TestNew_LabelExternalLinks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := NewPageData{
		CSRFToken: "test-csrf",
		FormInput: &viewmodel.DBWorkFormInput{
			OfficialSiteURL: "https://example.com",
			WikipediaURL:    "https://ja.wikipedia.org/wiki/x",
			TwitterUsername: "annict_com",
			TwitterHashtag:  "annict",
			ScTid:           "3524",
			MalAnimeID:      "20",
		},
	}

	var buf strings.Builder
	if err := New(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		// The external link opens in a new tab with tabnabbing protection and carries an
		// accessible name because the icon itself is decorative (aria-hidden).
		//
		// [Ja] 外部リンクは tabnabbing 対策付きで新しいタブで開き、アイコンは装飾 (aria-hidden) なので
		// リンク側にアクセシブルネームを持つ。
		`target="_blank"`,
		`rel="noopener"`,
		`aria-label="公式サイトURL を新しいタブで開く"`,
		// URL fields link to the submitted value itself; the id / username fields derive
		// their service URL via the shared helpers.
		//
		// [Ja] URL 系フィールドは送信値自体を、ID / ユーザー名系は共有ヘルパーで導出した URL をリンクする。
		`href="https://example.com"`,
		`href="https://ja.wikipedia.org/wiki/x"`,
		`href="https://x.com/annict_com"`,
		`href="http://cal.syoboi.jp/tid/3524"`,
		`href="https://myanimelist.net/anime/20"`,
		// synopsis / synopsis_en use 10 rows to match the Rails form.
		//
		// [Ja] synopsis / synopsis_en は Rails フォームに合わせて 10 行にする。
		`rows="10"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestNew_NoLabelExternalLinksWhenEmpty verifies that no external link is rendered when
// the linkable fields are empty.
//
// [Ja] TestNew_NoLabelExternalLinksWhenEmpty は、値が空のとき外部リンクが描画されないことを
// テストする。
func TestNew_NoLabelExternalLinksWhenEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := NewPageData{
		CSRFToken: "test-csrf",
		FormInput: &viewmodel.DBWorkFormInput{},
	}

	var buf strings.Builder
	if err := New(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "<form") {
		t.Error("フォームは描画されるべきです")
	}
	if strings.Contains(html, "を新しいタブで開く") {
		t.Error("値が空のとき外部リンクは描画されてはいけません")
	}
}

// TestNew_SidebarToggle verifies the new form renders the sidebar toggle in its header.
//
// [Ja] TestNew_SidebarToggle は新規フォームがヘッダーにサイドバートグルを描画する
// ことを検証する。
func TestNew_SidebarToggle(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	if err := New(NewPageData{FormInput: &viewmodel.DBWorkFormInput{}}).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	// The toggle is wired to the sidebar at every viewport size.
	//
	// [Ja] トグルはサイドバーに結線され、全画面幅で利用できる。
	for _, expected := range []string{`data-sidebar-toggle="db-sidebar"`} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestEdit_SidebarToggle verifies the edit form renders the sidebar toggle in its header.
//
// [Ja] TestEdit_SidebarToggle は編集フォームがヘッダーにサイドバートグルを描画する
// ことを検証する。
func TestEdit_SidebarToggle(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	if err := Edit(EditPageData{WorkID: 1, FormInput: &viewmodel.DBWorkFormInput{}}).Render(context.Background(), &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	// The toggle is wired to the sidebar at every viewport size.
	//
	// [Ja] トグルはサイドバーに結線され、全画面幅で利用できる。
	for _, expected := range []string{`data-sidebar-toggle="db-sidebar"`} {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}

// TestEdit_LabelExternalLinks verifies that the edit form renders external links through
// the shared sub-template too, confirming it is wired the same way as the new form.
//
// [Ja] TestEdit_LabelExternalLinks は、編集フォームでも共有サブテンプレート経由で外部リンクが
// 描画されることをテストする (新規フォームと同じ配線であることの担保)。
func TestEdit_LabelExternalLinks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := EditPageData{
		CSRFToken: "test-csrf",
		WorkID:    1,
		FormInput: &viewmodel.DBWorkFormInput{
			OfficialSiteURL: "https://example.com",
			ScTid:           "3524",
		},
	}

	var buf strings.Builder
	if err := Edit(data).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	expectedContents := []string{
		`aria-label="公式サイトURL を新しいタブで開く"`,
		`href="https://example.com"`,
		`href="http://cal.syoboi.jp/tid/3524"`,
		`rows="10"`,
	}
	for _, expected := range expectedContents {
		if !strings.Contains(html, expected) {
			t.Errorf("期待する文字列が含まれていません: %q", expected)
		}
	}
}
