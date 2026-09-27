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

// renderDbLayoutは指定したAccept-LanguageでDbレイアウトを描画し、
// 生成されたHTMLを返す。
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

// TestDb_RendersSidebarWithoutToggleはDbレイアウトがサイドバー本体を描画する一方、
// 開閉トグルはもう埋め込まないことを検証する。トグルは各ページのタイトル行に移った
// (components.DBSidebarToggleを参照) ため、レイアウトはトグルが参照するサイドバーのidを
// 公開するだけになる。
func TestDb_RendersSidebarWithoutToggle(t *testing.T) {
	t.Parallel()

	html := renderDbLayout(t, "ja")

	// ページに置かれたトグルが参照できるよう、サイドバーはidを公開する。
	if !strings.Contains(html, `<aside id="db-sidebar"`) {
		t.Error("レイアウトにサイドバー本体が描画されていません")
	}

	// トグルはもうレイアウトに属さず、各ページ側で描画される。
	if strings.Contains(html, `data-sidebar-toggle="db-sidebar"`) {
		t.Error("レイアウトにサイドバー開閉トグルが含まれてはいけません (ページ側へ移設済み)")
	}
}

// TestDb_HeadOmitsPublicPageMetaTagsはDbレイアウトが <head> をcomponents.DBHeadで
// 組み立てることを検証する。これにより検索エンジンとインストール済みPWAに向けたメタタグは
// 管理画面に出ず、共通のタグと最小限のOpen Graphは従来どおり描画される。
func TestDb_HeadOmitsPublicPageMetaTags(t *testing.T) {
	t.Parallel()

	html := renderDbLayout(t, "ja")

	// DefaultPageMetaはdescriptionとOG画像を埋めるため、レイアウトが公開ページ用の
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
			t.Errorf("/dbの <head> に公開ページ向けの出力が含まれてはいけません: %q", notWant)
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
			t.Errorf("/dbの <head> に共通の出力が含まれていません: %q", want)
		}
	}
}

// TestDb_FullHeightFollowsVisibleViewportは背面領域の高さが可視ビューポートに追随する
// ことを検証する。モバイルのツールバー表示時に領域が画面より高くなり、本来1画面に収まる
// ページにスクロールバーが出ることを防ぐ。
func TestDb_FullHeightFollowsVisibleViewport(t *testing.T) {
	t.Parallel()

	html := renderDbLayout(t, "ja")

	// class属性全体を固定するため、静的な単位へ戻せばここで落ちる。
	if !strings.Contains(html, `<div class="min-h-dvh">`) {
		t.Error("背面領域のフルハイト指定がmin-h-dvhになっていません")
	}
}

// dbMainIDはレイアウトの本文領域のidで、スキップリンクの飛び先でもある。下記の
// 検証はどちらもこの定数から組み立てるため、リンクと飛び先がずれることがない。
const dbMainID = "db-main"

// TestDb_RendersSkipLinkはDbレイアウトが本文領域へ飛ぶスキップリンクを両ロケールで
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
				t.Errorf("スキップリンクのラベル%qが描画されていません", tt.label)
			}

			// リンクは隠れているがフォーカス可能で (display: noneではなくsr-only)、
			// フォーカスされると表示される。
			if !strings.Contains(html, "sr-only focus:not-sr-only") {
				t.Error("スキップリンクが「既定では隠れ、フォーカス時に表示される」形になっていません")
			}

			// スキップリンクはDOM上でサイドバーより前にあって初めてサイドバーを飛ばせる。
			sidebarIndex := strings.Index(html, `<aside id="db-sidebar"`)
			if sidebarIndex == -1 {
				t.Fatal("レイアウトにサイドバー本体が描画されていません")
			}
			if skipLinkIndex > sidebarIndex {
				t.Error("スキップリンクがサイドバーより後ろに描画されています (サイドバーを飛ばせません)")
			}

			// 飛び先はリンクが指すidを持ち、飛んだときに実際にフォーカスが本文領域へ
			// 移るようtabindex="-1" を持つ。
			if !strings.Contains(html, `<main id="`+dbMainID+`" tabindex="-1">`) {
				t.Errorf("スキップリンクの飛び先 <main id=%q tabindex=\"-1\"> が描画されていません", dbMainID)
			}
		})
	}
}
