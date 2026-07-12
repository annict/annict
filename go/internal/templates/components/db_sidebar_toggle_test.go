package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/templates"
)

// TestDBSidebarToggle verifies the toggle renders a native disclosure button wired to the
// sidebar at every viewport size.
//
// [Ja] TestDBSidebarToggle はトグルが全画面幅でサイドバーに結線されるネイティブの
// disclosure ボタンを描画することを検証する。
func TestDBSidebarToggle(t *testing.T) {
	t.Parallel()

	ctx := i18n.SetLocale(context.Background(), "ja")

	var buf strings.Builder
	if err := DBSidebarToggle().Render(ctx, &buf); err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}
	html := buf.String()

	checks := []string{
		// A native button (keyboard operable) following the disclosure ARIA pattern:
		// aria-controls points at the sidebar and aria-expanded reflects the open state.
		// data-sidebar-toggle wires it to the sidebar via sidebar-toggle.ts.
		//
		// [Ja] disclosure の ARIA パターンに従うネイティブ button (キーボード操作可)。
		// aria-controls がサイドバーを指し、aria-expanded が開閉状態を表す。
		// data-sidebar-toggle が sidebar-toggle.ts 経由でサイドバーに結線する。
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

// TestDBSidebarCloseButton verifies the sidebar has a clearly labelled native close button.
//
// [Ja] TestDBSidebarCloseButton はサイドバー内に明確なラベル付きのネイティブな閉じる
// ボタンがあることを検証する。
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
				`data-sidebar-close="db-sidebar"`,
				`aria-controls="db-sidebar"`,
				tt.label,
			} {
				if !strings.Contains(html, expected) {
					t.Errorf("HTMLに必要な要素が含まれていません: %q", expected)
				}
			}
		})
	}
}

// TestDBSidebarInitialState verifies the server-rendered sidebar and toggle reflect the desktop
// preference before client JavaScript runs.
//
// [Ja] TestDBSidebarInitialState はクライアント JavaScript の実行前から、SSR されたサイドバーと
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
			name:             "open",
			open:             true,
			sidebarChecks:    []string{`data-desktop-open="true"`, `aria-hidden="false"`},
			sidebarNotChecks: []string{`data-initial-open="false"`, ` inert`},
			toggleCheck:      `aria-expanded="true"`,
		},
		{
			name:          "closed",
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
					t.Errorf("サイドバー HTML に必要な初期状態が含まれていません: %q", expected)
				}
			}
			for _, unexpected := range tt.sidebarNotChecks {
				if strings.Contains(sidebarBuf.String(), unexpected) {
					t.Errorf("サイドバー HTML に不要な初期状態が含まれています: %q", unexpected)
				}
			}

			var toggleBuf strings.Builder
			if err := DBSidebarToggle().Render(ctx, &toggleBuf); err != nil {
				t.Fatalf("トグルのレンダリングエラー: %v", err)
			}
			if !strings.Contains(toggleBuf.String(), tt.toggleCheck) {
				t.Errorf("トグル HTML に必要な初期状態が含まれていません: %q", tt.toggleCheck)
			}
		})
	}
}

// TestDBSidebarToggleI18n verifies the toggle's aria-label switches per locale.
//
// [Ja] TestDBSidebarToggleI18n はトグルの aria-label が言語ごとに切り替わることを検証する。
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
				t.Errorf("トグルの aria-label が正しくありません: 期待=%s", tt.label)
			}
		})
	}
}
