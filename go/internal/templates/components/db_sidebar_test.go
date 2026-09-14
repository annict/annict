package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates"
)

// renderDBSidebarは指定したリクエストパスでDBサイドバーを描画し、生成されたHTMLを
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

// TestDBSidebar_MarksCurrentPageは、表示中の画面の項目にaria-currentが付き塗りつぶ
// されること (一覧パスでも配下のパスでも) と、他の項目には付かないことを検証する。作品の編集
// ページでも作品の項目に印が残り、接頭辞の長い /db/series_worksが /db/seriesに染み出さない。
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
				t.Errorf("%qの項目にaria-currentが付いていません", tt.wantCurrent)
			}
			// 塗りつぶしは同じ属性で切り替わるため、目に見える印が支援技術の読む印から
			// ずれることはない。
			for _, expected := range []string{
				`aria-[current=page]:bg-primary`,
				`aria-[current=page]:text-primary-foreground`,
				`aria-[current=page]:hover:bg-primary`,
			} {
				if !strings.Contains(current, expected) {
					t.Errorf("%qの項目に強調クラスが含まれていません: %q", tt.wantCurrent, expected)
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
					t.Errorf("現在ページではない%qの項目にaria-currentが付いています", path)
				}
			}
		})
	}
}

// TestDBSidebar_MarksNoEntryOutsideMenuは、サイドバーに項目を持たない画面 (DBの検索
// 結果) ではどの項目にも印が付かないことを検証する。
func TestDBSidebar_MarksNoEntryOutsideMenu(t *testing.T) {
	t.Parallel()

	html := renderDBSidebar(t, "/db/search")

	if strings.Contains(html, `aria-current="page"`) {
		t.Error("サイドバーに項目を持たない画面でaria-currentが付いています")
	}
}

// TestDBSidebar_RendersBackLinkInFooterは、戻るリンクがスクロールするsectionではなく
// footerに描画されることを検証する。これがメニューがスクロールするときに戻るリンクを
// サイドバーの下端へ残す仕組みである。
func TestDBSidebar_RendersBackLinkInFooter(t *testing.T) {
	t.Parallel()

	html := renderDBSidebar(t, "/db/works")

	footerStart := strings.Index(html, "<footer")
	if footerStart < 0 {
		t.Fatal("footerが描画されていません")
	}
	footerEnd := strings.Index(html[footerStart:], "</footer>")
	if footerEnd < 0 {
		t.Fatal("footerが閉じられていません")
	}
	footer := html[footerStart : footerStart+footerEnd]

	if !strings.Contains(footer, `<a href="/"`) {
		t.Error("戻るリンクがfooterの内側に描画されていません")
	}
	if !strings.Contains(footer, "Annictに戻る") {
		t.Error("戻るリンクのラベルがfooterの内側に描画されていません")
	}
	// リンクはfooterの幅いっぱいに広がるため、アイコンと文言をサイドバーの横中央に
	// 置くのは中央寄せの指定である。
	if !strings.Contains(footer, "justify-center") {
		t.Error("戻るリンクが中央寄せになっていません")
	}

	sectionEnd := strings.Index(html, "</section>")
	if sectionEnd < 0 {
		t.Fatal("sectionが閉じられていません")
	}
	if sectionEnd > footerStart {
		t.Error("footerがsectionより前に描画されています")
	}
	if strings.Contains(html[:sectionEnd], `<a href="/"`) {
		t.Error("戻るリンクがsectionの内側に残っています")
	}
}

// TestDBSidebar_AlignsHorizontalPaddingは、検索欄がBasecoatがp-2で字下げする
// メニューのグループとfooterと同じ横方向の余白を取ることを検証する。ここを揃えることで、
// 検索欄・メニュー項目・戻るリンクの左端が縦に一直線に並ぶ。
func TestDBSidebar_AlignsHorizontalPadding(t *testing.T) {
	t.Parallel()

	html := renderDBSidebar(t, "/db/works")

	if !strings.Contains(html, `<div class="px-2">`) {
		t.Error("検索欄のラッパーがpx-2になっていません")
	}
	if strings.Contains(html, `class="px-4"`) {
		t.Error("検索欄のラッパーにpx-4が残っています")
	}
}
