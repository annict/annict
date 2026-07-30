package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates"
)

// renderDBSidebar renders the DB sidebar for the given request path and returns the
// produced HTML.
//
// [Ja] renderDBSidebar は指定したリクエストパスで DB サイドバーを描画し、生成された HTML を
// 返す。
func renderDBSidebar(t *testing.T, currentPath string) string {
	t.Helper()

	ctx := templates.SetCurrentPath(i18n.SetLocale(context.Background(), "ja"), currentPath)

	var buf strings.Builder
	if err := DBSidebar().Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	return buf.String()
}

// menuItemHTML returns the markup of the sidebar entry linking to the given path.
//
// [Ja] menuItemHTML は指定したパスへリンクするサイドバー項目のマークアップを返す。
func menuItemHTML(t *testing.T, html string, path string) string {
	t.Helper()

	start := strings.Index(html, `<a href="`+path+`"`)
	if start < 0 {
		t.Fatalf("%q へのリンクが描画されていません", path)
	}
	end := strings.Index(html[start:], "</a>")
	if end < 0 {
		t.Fatalf("%q へのリンクが閉じられていません", path)
	}
	return html[start : start+end]
}

// TestDBSidebar_MarksCurrentPage verifies the entry of the screen being viewed is marked
// with aria-current and filled, both on the index path and on a path below it, and that no
// other entry is marked. A work edit page keeps the works entry marked, and the entry with
// the longer prefix (/db/series_works) does not bleed into /db/series.
//
// [Ja] TestDBSidebar_MarksCurrentPage は、表示中の画面の項目に aria-current が付き塗りつぶ
// されること (一覧パスでも配下のパスでも) と、他の項目には付かないことを検証する。作品の編集
// ページでも作品の項目に印が残り、接頭辞の長い /db/series_works が /db/series に染み出さない。
func TestDBSidebar_MarksCurrentPage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		currentPath string
		wantCurrent string
	}{
		{
			name:        "一覧パス",
			currentPath: "/db/works",
			wantCurrent: "/db/works",
		},
		{
			name:        "編集ページ",
			currentPath: "/db/works/1/edit",
			wantCurrent: "/db/works",
		},
		{
			name:        "新規作成ページ",
			currentPath: "/db/works/new",
			wantCurrent: "/db/works",
		},
		{
			name:        "他の画面",
			currentPath: "/db/characters",
			wantCurrent: "/db/characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			html := renderDBSidebar(t, tt.currentPath)

			current := menuItemHTML(t, html, tt.wantCurrent)
			if !strings.Contains(current, `aria-current="page"`) {
				t.Errorf("%q の項目に aria-current が付いていません", tt.wantCurrent)
			}
			// The fill is driven by the same attribute, so the visible marker cannot drift
			// away from the one assistive technology reads.
			//
			// [Ja] 塗りつぶしは同じ属性で切り替わるため、目に見える印が支援技術の読む印から
			// ずれることはない。
			for _, expected := range []string{
				`aria-[current=page]:bg-primary`,
				`aria-[current=page]:text-primary-foreground`,
				`aria-[current=page]:hover:bg-primary`,
			} {
				if !strings.Contains(current, expected) {
					t.Errorf("%q の項目に強調クラスが含まれていません: %q", tt.wantCurrent, expected)
				}
			}

			for _, path := range []string{
				"/db/activities",
				"/db/series",
				"/db/works",
				"/db/people",
				"/db/organizations",
				"/db/characters",
				"/db/channel_groups",
				"/db/channels",
			} {
				if path == tt.wantCurrent {
					continue
				}
				if strings.Contains(menuItemHTML(t, html, path), `aria-current="page"`) {
					t.Errorf("現在ページではない %q の項目に aria-current が付いています", path)
				}
			}
		})
	}
}

// TestDBSidebar_MarksNoEntryOutsideMenu verifies a screen that has no sidebar entry (the DB
// search results) leaves every entry unmarked.
//
// [Ja] TestDBSidebar_MarksNoEntryOutsideMenu は、サイドバーに項目を持たない画面 (DB の検索
// 結果) ではどの項目にも印が付かないことを検証する。
func TestDBSidebar_MarksNoEntryOutsideMenu(t *testing.T) {
	t.Parallel()

	html := renderDBSidebar(t, "/db/search")

	if strings.Contains(html, `aria-current="page"`) {
		t.Error("サイドバーに項目を持たない画面で aria-current が付いています")
	}
}

// TestDBSidebar_RendersBackLinkInFooter verifies the back link is rendered in the footer
// instead of the scrolling section, which is what keeps it at the bottom edge of the sidebar
// when the menu scrolls.
//
// [Ja] TestDBSidebar_RendersBackLinkInFooter は、戻るリンクがスクロールする section ではなく
// footer に描画されることを検証する。これがメニューがスクロールするときに戻るリンクを
// サイドバーの下端へ残す仕組みである。
func TestDBSidebar_RendersBackLinkInFooter(t *testing.T) {
	t.Parallel()

	html := renderDBSidebar(t, "/db/works")

	footerStart := strings.Index(html, "<footer")
	if footerStart < 0 {
		t.Fatal("footer が描画されていません")
	}
	footerEnd := strings.Index(html[footerStart:], "</footer>")
	if footerEnd < 0 {
		t.Fatal("footer が閉じられていません")
	}
	footer := html[footerStart : footerStart+footerEnd]

	if !strings.Contains(footer, `<a href="/"`) {
		t.Error("戻るリンクが footer の内側に描画されていません")
	}
	if !strings.Contains(footer, "Annictに戻る") {
		t.Error("戻るリンクのラベルが footer の内側に描画されていません")
	}
	// The link fills the footer's width, so centering is what puts its icon and label in the
	// middle of the sidebar.
	//
	// [Ja] リンクは footer の幅いっぱいに広がるため、アイコンと文言をサイドバーの横中央に
	// 置くのは中央寄せの指定である。
	if !strings.Contains(footer, "justify-center") {
		t.Error("戻るリンクが中央寄せになっていません")
	}

	sectionEnd := strings.Index(html, "</section>")
	if sectionEnd < 0 {
		t.Fatal("section が閉じられていません")
	}
	if sectionEnd > footerStart {
		t.Error("footer が section より前に描画されています")
	}
	if strings.Contains(html[:sectionEnd], `<a href="/"`) {
		t.Error("戻るリンクが section の内側に残っています")
	}
}

// TestDBSidebar_AlignsHorizontalPadding verifies the search box takes the same horizontal
// padding as the menu group and the footer, both of which Basecoat indents by p-2. Aligning
// it here keeps the left edges of the search field, the menu entries and the back link on one
// vertical line.
//
// [Ja] TestDBSidebar_AlignsHorizontalPadding は、検索欄が Basecoat が p-2 で字下げする
// メニューのグループと footer と同じ横方向の余白を取ることを検証する。ここを揃えることで、
// 検索欄・メニュー項目・戻るリンクの左端が縦に一直線に並ぶ。
func TestDBSidebar_AlignsHorizontalPadding(t *testing.T) {
	t.Parallel()

	html := renderDBSidebar(t, "/db/works")

	if !strings.Contains(html, `<div class="px-2">`) {
		t.Error("検索欄のラッパーが px-2 になっていません")
	}
	if strings.Contains(html, `class="px-4"`) {
		t.Error("検索欄のラッパーに px-4 が残っています")
	}
}
