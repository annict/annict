package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

// TestFooter_LegalDocumentLinksは、フッターの文書リンクがリクエストのロケールの言語版を指し、
// その言語版が無い文書では日本語版を指すことを検証する。
func TestFooter_LegalDocumentLinks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		locale      string
		expected    []string
		notExpected []string
	}{
		{
			name:   "日本語",
			locale: i18n.LangJa,
			expected: []string{
				`href="/terms">利用規約</a>`,
				`href="/privacy">プライバシーポリシー</a>`,
				`href="/legal">特商法表記</a>`,
			},
			notExpected: []string{
				"hreflang",
			},
		},
		{
			// 英語版を一覧表に登録した文書は英語版を指す。
			// 英語版を作らない文書は英語を使う利用者にも日本語版を指し、
			// リンク先がページと異なる言語であることをhreflangで示す
			name:   "英語",
			locale: i18n.LangEn,
			expected: []string{
				`href="/en/terms">Terms of Service</a>`,
				`href="/en/privacy">Privacy Policy</a>`,
				`href="/legal" hreflang="ja">Legal Notice</a>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			var buf strings.Builder
			if err := Footer(ctx).Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()
			for _, expected := range tt.expected {
				if !strings.Contains(html, expected) {
					t.Errorf("期待する文字列が含まれていません: %q", expected)
				}
			}
			for _, notExpected := range tt.notExpected {
				if strings.Contains(html, notExpected) {
					t.Errorf("含まれてはいけない文字列が含まれています: %q", notExpected)
				}
			}
		})
	}
}
