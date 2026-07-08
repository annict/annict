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

	meta := viewmodel.DefaultPageMeta(ctx, cfg)

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

// TestDb_SidebarToggle verifies the sidebar toggle button is rendered in the Db layout.
//
// [Ja] TestDb_SidebarToggle は Db レイアウトにサイドバー開閉トグルが描画されることを確認する。
func TestDb_SidebarToggle(t *testing.T) {
	t.Parallel()

	html := renderDbLayout(t, "ja")

	checks := []string{
		// The sidebar exposes an id so the toggle can target it.
		//
		// [Ja] トグルが参照できるよう、サイドバーは id を公開する。
		`<aside id="db-sidebar"`,
		// The toggle is a native button (keyboard operable) with the disclosure
		// ARIA pattern: aria-controls points at the sidebar and aria-expanded
		// reflects the open state. data-sidebar-toggle wires it to the sidebar
		// via the sidebar-toggle.ts module (the behavior lives in JS, not here).
		//
		// [Ja] トグルはネイティブ button (キーボード操作可) で、disclosure の ARIA
		// パターンを持つ。aria-controls がサイドバーを指し、aria-expanded が開閉状態
		// を表す。data-sidebar-toggle が sidebar-toggle.ts モジュール経由でサイドバー
		// に結線する (挙動は JS 側にあり、ここには無い)。
		`type="button"`,
		`data-sidebar-toggle="db-sidebar"`,
		`aria-controls="db-sidebar"`,
		`aria-expanded="true"`,
		`aria-label="サイドバーの開閉"`,
	}

	for _, expected := range checks {
		if !strings.Contains(html, expected) {
			t.Errorf("HTMLに必要な要素が含まれていません: %q", expected)
		}
	}
}

// TestDb_SidebarToggleI18n verifies the toggle's aria-label switches per locale.
//
// [Ja] TestDb_SidebarToggleI18n はトグルの aria-label が言語ごとに切り替わることを確認する。
func TestDb_SidebarToggleI18n(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		acceptLanguage string
		label          string
	}{
		{"日本語", "ja", `aria-label="サイドバーの開閉"`},
		{"英語", "en", `aria-label="Toggle sidebar"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			html := renderDbLayout(t, tt.acceptLanguage)
			if !strings.Contains(html, tt.label) {
				t.Errorf("トグルの aria-label が正しくありません: 期待=%s", tt.label)
			}
		})
	}
}
