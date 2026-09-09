package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
	"github.com/annict/annict/go/internal/viewmodel"
)

// TestSidebar_RendersDeveloperHelpInMiscGroup verifies the sidebar renders the Developer Help
// link once in the miscellaneous group with the exact published URL and localized label.
//
// [Ja] TestSidebar_RendersDeveloperHelpInMiscGroupは、サイドバーが開発者向けヘルプのリンクを
// 公開済みの完全なURLとローカライズ済みの表示名で「その他」グループに1つだけ描画することを
// 検証する。
func TestSidebar_RendersDeveloperHelpInMiscGroup(t *testing.T) {
	t.Parallel()

	const (
		developerHelpURL       = "https://wikino.app/s/annict/topics/5"
		legacyDeveloperHelpURL = "https://developers.annict.com/"
	)

	tests := []struct {
		name      string
		locale    string
		wantLabel string
	}{
		{
			name:      "日本語",
			locale:    "ja",
			wantLabel: "開発者向けヘルプ",
		},
		{
			name:      "英語",
			locale:    "en",
			wantLabel: "Developer Help",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			var buf strings.Builder
			if err := Sidebar(ctx, nil, viewmodel.Seasons{}).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}
			html := buf.String()

			if count := strings.Count(html, `href="`+developerHelpURL+`"`); count != 1 {
				t.Errorf("開発者向けヘルプのリンク数 = %d, want 1", count)
			}

			miscGroup := sidebarGroupHTML(t, html, "sidebar-misc")
			developerHelpLink := menuItemHTML(t, miscGroup, developerHelpURL)
			if !strings.Contains(developerHelpLink, tt.wantLabel) {
				t.Errorf("開発者向けヘルプのリンクに表示名 %q が含まれていません", tt.wantLabel)
			}

			servicesGroup := sidebarGroupHTML(t, html, "sidebar-services")
			if strings.Contains(servicesGroup, developerHelpURL) {
				t.Error("開発者向けヘルプのリンクが「サービス」グループに残っています")
			}

			if strings.Contains(html, legacyDeveloperHelpURL) {
				t.Error("旧Annict DevelopersのURLがサイドバーに残っています")
			}
		})
	}
}
