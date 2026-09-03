package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

// TestGuidelineLink verifies that the link points at the help page it was given, opens it in
// a new tab with tabnabbing protection, and says so in its accessible name in both locales.
// The visible text is part of that name, so the name never drops what the link shows.
//
// [Ja] TestGuidelineLink は、リンクが渡されたヘルプページを指し、tabnabbing 対策付きで新しい
// タブに開き、そのことを両ロケールのアクセシブルネームで伝えることを検証する。可視テキストは
// アクセシブルネームに含まれるため、リンクの表示内容が名前から落ちることはない。
func TestGuidelineLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		locale        string
		label         string
		wantAriaLabel string
	}{
		{
			name:          "ja",
			locale:        "ja",
			label:         "作品の編集ガイドライン",
			wantAriaLabel: `aria-label="作品の編集ガイドライン を新しいタブで開く"`,
		},
		{
			name:          "en",
			locale:        "en",
			label:         "Work editing guidelines",
			wantAriaLabel: `aria-label="Open Work editing guidelines in a new tab"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			var buf strings.Builder
			if err := GuidelineLink(tt.label, "https://example.com/pages/1").Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			for _, expected := range []string{
				`href="https://example.com/pages/1"`,
				`target="_blank"`,
				`rel="noopener"`,
				tt.wantAriaLabel,
				tt.label,
				// The icon repeats what the link already says, so it stays out of the
				// accessibility tree and out of the focus order.
				//
				// [Ja] アイコンはリンクが既に伝えている内容を繰り返すだけなので、
				// アクセシビリティツリーとフォーカス順序から外す。
				`aria-hidden="true"`,
				`focusable="false"`,
			} {
				if !strings.Contains(html, expected) {
					t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", expected, html)
				}
			}
		})
	}
}
