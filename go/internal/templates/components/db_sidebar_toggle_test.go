package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates"
)

// TestDBSidebarToggleはトグルが全画面幅でサイドバーに結線されるネイティブの
// disclosureボタンを描画することを検証する。
func TestDBSidebarToggle(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := DBSidebarToggle().Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	checks := []string{
		// disclosureのARIAパターンに従うネイティブbutton (キーボード操作可)。
		// aria-controlsがサイドバーを指し、aria-expandedが開閉状態を表す。
		// data-sidebar-toggleがsidebar-toggle.ts経由でサイドバーに結線する。
		`type="button"`,
		`data-variant="outline"`,
		`data-size="icon"`,
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
	if strings.Contains(html, "md:hidden") {
		t.Error("サイドバー開閉トグルはデスクトップでも表示される必要があります")
	}
}

// TestDBSidebarToggle_DecorativeIconIsHiddenはトグルのアイコンがアクセシビリティツリー
// とフォーカス順序から外れることを検証する。ボタンはaria-labelで既に意味を伝えるため、SVGは
// ブラウザー依存の別表現を重ねるだけになる。隠すのはラッパー要素ではなくSVG自体で行う。
// ラッパーでは、SVG要素を既定でフォーカス可能とする実装でSVGがフォーカス可能なまま残るため。
// ボタンにテキストが無く添える相手がいないので、Basecoatのinline位置は指定しない。
func TestDBSidebarToggle_DecorativeIconIsHidden(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var iconBuf strings.Builder
	if err := templates.DecorativeIcon("sidebar-regular").Render(ctx, &iconBuf); err != nil {
		t.Fatalf("アイコンのレンダリングエラー: %v", err)
	}

	var buf strings.Builder
	if err := DBSidebarToggle().Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	if !strings.Contains(html, iconBuf.String()) {
		t.Error(`装飾アイコン "sidebar-regular" がaria-hiddenかつfocusable="false" ではありません`)
	}
	if strings.Contains(html, `<span aria-hidden="true">`) {
		t.Error("装飾アイコンはラッパー要素ではなくSVG自体で隠すべきです")
	}
}

// TestDBSidebarCloseButtonはサイドバー内に、アクセシブルな名前を持つモバイル専用
// (md:hidden) のアイコンだけのネイティブな閉じるボタンがあることを検証する。デスクトップでは
// 常時表示のサイドバートグルで既に閉じられるため、このサイドバー内ボタンは非表示になる。
func TestDBSidebarCloseButton(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		label  string
	}{
		{"日本語", "ja", "閉じる"},
		{"英語", "en", "Close"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)
			var buf strings.Builder
			if err := DBSidebar().Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}
			html := buf.String()
			for _, expected := range []string{
				`type="button"`,
				`data-variant="ghost"`,
				`data-size="icon"`,
				`data-sidebar-close="db-sidebar"`,
				`aria-controls="db-sidebar"`,
				`aria-label="` + tt.label + `"`,
				`M205.66,194.34a8,8,0,0,1-11.32,11.32`,
				// モバイル専用: md以上ではサイドバートグルが閉じるため非表示。
				`md:hidden`,
			} {
				if !strings.Contains(html, expected) {
					t.Errorf("HTMLに必要な要素が含まれていません: %q", expected)
				}
			}
		})
	}
}

// TestDBSidebarInitialStateはクライアントJavaScriptの実行前から、SSRされたサイドバーと
// トグルへデスクトップ設定が反映されることを検証する。
func TestDBSidebarInitialState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		open             bool
		sidebarChecks    []string
		sidebarNotChecks []string
		toggleCheck      string
	}{
		{
			name:             "開いている状態",
			open:             true,
			sidebarChecks:    []string{`data-desktop-open="true"`, `aria-hidden="false"`},
			sidebarNotChecks: []string{`data-initial-open="false"`, ` inert`},
			toggleCheck:      `aria-expanded="true"`,
		},
		{
			name:          "閉じている状態",
			open:          false,
			sidebarChecks: []string{`data-desktop-open="false"`, `aria-hidden="true"`, `data-initial-open="false"`, ` inert`},
			toggleCheck:   `aria-expanded="false"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := templates.SetDBSidebarOpen(i18n.SetLocale(context.Background(), "en"), tt.open)
			var sidebarBuf strings.Builder
			if err := DBSidebar().Render(ctx, &sidebarBuf); err != nil {
				t.Fatalf("サイドバーのレンダリングエラー: %v", err)
			}
			for _, expected := range tt.sidebarChecks {
				if !strings.Contains(sidebarBuf.String(), expected) {
					t.Errorf("サイドバーHTMLに必要な初期状態が含まれていません: %q", expected)
				}
			}
			for _, unexpected := range tt.sidebarNotChecks {
				if strings.Contains(sidebarBuf.String(), unexpected) {
					t.Errorf("サイドバーHTMLに不要な初期状態が含まれています: %q", unexpected)
				}
			}

			var toggleBuf strings.Builder
			if err := DBSidebarToggle().Render(ctx, &toggleBuf); err != nil {
				t.Fatalf("トグルのレンダリングエラー: %v", err)
			}
			if !strings.Contains(toggleBuf.String(), tt.toggleCheck) {
				t.Errorf("トグルHTMLに必要な初期状態が含まれていません: %q", tt.toggleCheck)
			}
		})
	}
}

// TestDBSidebarToggleI18nはトグルのaria-labelが言語ごとに切り替わることを検証する。
func TestDBSidebarToggleI18n(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		label  string
	}{
		{"日本語", "ja", `aria-label="サイドバーの開閉"`},
		{"英語", "en", `aria-label="Toggle sidebar"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			var buf strings.Builder
			if err := DBSidebarToggle().Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}
			if !strings.Contains(buf.String(), tt.label) {
				t.Errorf("トグルのaria-labelが正しくありません: 期待=%s", tt.label)
			}
		})
	}
}
