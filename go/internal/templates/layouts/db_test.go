package layouts

import (
	"bytes"
	"context"
	"io"
	"net/http"
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

	var ctx context.Context
	i18nHandler := i18n.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
	}))
	i18nHandler.ServeHTTP(httptest.NewRecorder(), req)

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
