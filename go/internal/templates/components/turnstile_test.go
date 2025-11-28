package components

import (
	"context"
	"strings"
	"testing"
)

func TestTurnstile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		siteKey      string
		wantContains []string
		wantNotFound []string
	}{
		{
			name:    "Site Keyが設定されている場合、Turnstileウィジェットが表示される",
			siteKey: "1x00000000000000000000AA",
			wantContains: []string{
				`<script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>`,
				`<div class="cf-turnstile"`,
				`data-sitekey="1x00000000000000000000AA"`,
			},
		},
		{
			name:    "Site Keyが空の場合、何も表示されない",
			siteKey: "",
			wantNotFound: []string{
				`<script src="https://challenges.cloudflare.com/turnstile/v0/api.js"`,
				`<div class="cf-turnstile"`,
				`data-sitekey`,
				`data-appearance`,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // キャプチャ
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// テンプレートをレンダリング
			var buf strings.Builder
			err := Turnstile(tt.siteKey).Render(ctx, &buf)
			if err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			// 期待する文字列が含まれているか確認
			for _, want := range tt.wantContains {
				if !strings.Contains(html, want) {
					t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", want, html)
				}
			}

			// 含まれていないべき文字列が含まれていないか確認
			for _, notWant := range tt.wantNotFound {
				if strings.Contains(html, notWant) {
					t.Errorf("含まれるべきでない文字列が含まれています: %q\nHTML: %s", notWant, html)
				}
			}
		})
	}
}

func TestTurnstile_JavaScriptLoading(t *testing.T) {
	t.Parallel()

	siteKey := "test-site-key"
	ctx := context.Background()

	var buf strings.Builder
	err := Turnstile(siteKey).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// JavaScript が非同期で読み込まれるか確認
	if !strings.Contains(html, "async") {
		t.Error("Turnstile JavaScript が async 属性で読み込まれていません")
	}

	if !strings.Contains(html, "defer") {
		t.Error("Turnstile JavaScript が defer 属性で読み込まれていません")
	}
}

func TestTurnstile_InvisibleMode(t *testing.T) {
	t.Parallel()

	siteKey := "test-site-key"
	ctx := context.Background()

	var buf strings.Builder
	err := Turnstile(siteKey).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// data-appearance属性が含まれていないことを確認
	if strings.Contains(html, `data-appearance`) {
		t.Error("Turnstile に data-appearance 属性が含まれています")
	}
}
