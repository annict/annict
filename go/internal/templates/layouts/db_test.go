package layouts

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"github.com/annict/annict/go/internal/config"
	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// renderDbLayout renders the Db layout with the given Accept-Language and
// returns the produced HTML.
//
// [Ja] renderDbLayout は指定した Accept-Language で Db レイアウトを描画し、
// 生成された HTML を返す。
func renderDbLayout(t *testing.T, acceptLanguage string) string {
	t.Helper()

	cfg := &config.Config{
		Env:    "test",
		Domain: "annict.test",
	}

	req := httptest.NewRequest("GET", "/db/works", nil)
	req.Header.Set("Accept-Language", acceptLanguage)

	ctx := i18n.SetLocale(req.Context(), i18n.DetectLanguage(req))

	meta := viewmodel.DefaultPageMeta(ctx, cfg, req.URL.Path)

	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte("<div>Content</div>"))
		return err
	})

	var buf bytes.Buffer
	if err := Db(meta, "v1.0.0", content).Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	return buf.String()
}

// TestDb_RendersSidebarWithoutToggle verifies the Db layout renders the sidebar itself but
// no longer embeds the open/close toggle: the toggle now lives in each page's title row
// (see components.DBSidebarToggle), so the layout only exposes the sidebar's id for it.
//
// [Ja] TestDb_RendersSidebarWithoutToggle は Db レイアウトがサイドバー本体を描画する一方、
// 開閉トグルはもう埋め込まないことを検証する。トグルは各ページのタイトル行に移った
// (components.DBSidebarToggle を参照) ため、レイアウトはトグルが参照するサイドバーの id を
// 公開するだけになる。
func TestDb_RendersSidebarWithoutToggle(t *testing.T) {
	t.Parallel()

	html := renderDbLayout(t, "ja")

	// The sidebar exposes an id so a toggle placed on the page can target it.
	//
	// [Ja] ページに置かれたトグルが参照できるよう、サイドバーは id を公開する。
	if !strings.Contains(html, `<aside id="db-sidebar"`) {
		t.Error("レイアウトにサイドバー本体が描画されていません")
	}

	// The toggle no longer belongs to the layout; it is rendered per page instead.
	//
	// [Ja] トグルはもうレイアウトに属さず、各ページ側で描画される。
	if strings.Contains(html, `data-sidebar-toggle="db-sidebar"`) {
		t.Error("レイアウトにサイドバー開閉トグルが含まれてはいけません (ページ側へ移設済み)")
	}
}

// TestDb_HeadOmitsPublicPageMetaTags verifies the Db layout builds its <head> from
// components.DBHead, so the meta tags addressing search engines and the installed PWA stay
// out of the admin pages while the shared tags and the minimal Open Graph set keep
// rendering.
//
// [Ja] TestDb_HeadOmitsPublicPageMetaTags は Db レイアウトが <head> を components.DBHead で
// 組み立てることを検証する。これにより検索エンジンとインストール済み PWA に向けたメタタグは
// 管理画面に出ず、共通のタグと最小限の Open Graph は従来どおり描画される。
func TestDb_HeadOmitsPublicPageMetaTags(t *testing.T) {
	t.Parallel()

	html := renderDbLayout(t, "ja")

	// DefaultPageMeta fills in the description and the OG image, so these tags would be
	// rendered with real values if the layout still used the public <head>.
	//
	// [Ja] DefaultPageMeta は description と OG 画像を埋めるため、レイアウトが公開ページ用の
	// <head> を使い続けていればこれらのタグは実際の値付きで描画される。
	wantNotContains := []string{
		`name="description"`,
		`property="og:description"`,
		`property="og:image"`,
		`property="og:locale"`,
		`name="twitter:`,
		`rel="canonical"`,
		`rel="manifest"`,
	}
	for _, notWant := range wantNotContains {
		if strings.Contains(html, notWant) {
			t.Errorf("/db の <head> に公開ページ向けの出力が含まれてはいけません: %q", notWant)
		}
	}

	wantContains := []string{
		`<meta charset="UTF-8">`,
		`<title>`,
		`<link rel="stylesheet" href="/static/css/style.css?v=v1.0.0">`,
		`<script type="module" src="/static/js/main.js?v=v1.0.0"></script>`,
		`<meta property="og:title"`,
		`<meta property="og:type" content="website">`,
		`<meta property="og:site_name" content="Annict (アニクト)">`,
	}
	for _, want := range wantContains {
		if !strings.Contains(html, want) {
			t.Errorf("/db の <head> に共通の出力が含まれていません: %q", want)
		}
	}
}

// dbMainID is the id of the layout's main content region and the target of its skip link.
// Both assertions below derive from it, so the link and its destination cannot drift apart.
//
// [Ja] dbMainID はレイアウトの本文領域の id で、スキップリンクの飛び先でもある。下記の
// 検証はどちらもこの定数から組み立てるため、リンクと飛び先がずれることがない。
const dbMainID = "db-main"

// TestDb_RendersSkipLink verifies the Db layout offers a skip link that jumps to the main
// content region, in both locales. The link precedes the sidebar so it is the first thing a
// keyboard user reaches, and it is visually hidden until focused.
//
// [Ja] TestDb_RendersSkipLink は Db レイアウトが本文領域へ飛ぶスキップリンクを両ロケールで
// 提供することを検証する。リンクはキーボード利用者が最初に到達するようサイドバーより前に
// あり、フォーカスされるまで視覚的に隠れている。
func TestDb_RendersSkipLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		acceptLanguage string
		label          string
	}{
		{name: "日本語", acceptLanguage: "ja", label: "メインコンテンツへスキップ"},
		{name: "英語", acceptLanguage: "en", label: "Skip to main content"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			html := renderDbLayout(t, tt.acceptLanguage)

			skipLinkIndex := strings.Index(html, `href="#`+dbMainID+`"`)
			if skipLinkIndex == -1 {
				t.Fatalf("本文 (#%s) へのスキップリンクが描画されていません", dbMainID)
			}

			if !strings.Contains(html, tt.label) {
				t.Errorf("スキップリンクのラベル %q が描画されていません", tt.label)
			}

			// The link is hidden but focusable (sr-only, not display: none) and becomes
			// visible once focused.
			//
			// [Ja] リンクは隠れているがフォーカス可能で (display: none ではなく sr-only)、
			// フォーカスされると表示される。
			if !strings.Contains(html, "sr-only focus:not-sr-only") {
				t.Error("スキップリンクが「既定では隠れ、フォーカス時に表示される」形になっていません")
			}

			// The skip link only bypasses the sidebar if it comes before it in the DOM.
			//
			// [Ja] スキップリンクは DOM 上でサイドバーより前にあって初めてサイドバーを飛ばせる。
			sidebarIndex := strings.Index(html, `<aside id="db-sidebar"`)
			if sidebarIndex == -1 {
				t.Fatal("レイアウトにサイドバー本体が描画されていません")
			}
			if skipLinkIndex > sidebarIndex {
				t.Error("スキップリンクがサイドバーより後ろに描画されています (サイドバーを飛ばせません)")
			}

			// The destination carries the id the link points at, and tabindex="-1" so the
			// jump actually moves focus into the content region.
			//
			// [Ja] 飛び先はリンクが指す id を持ち、飛んだときに実際にフォーカスが本文領域へ
			// 移るよう tabindex="-1" を持つ。
			if !strings.Contains(html, `<main id="`+dbMainID+`" tabindex="-1">`) {
				t.Errorf("スキップリンクの飛び先 <main id=%q tabindex=\"-1\"> が描画されていません", dbMainID)
			}
		})
	}
}
