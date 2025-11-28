package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/internal/i18n"
	"github.com/annict/annict/internal/session"
)

func TestFlash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		flash         *session.Flash
		wantContains  []string
		wantNotRender bool
	}{
		{
			name: "成功メッセージ",
			flash: &session.Flash{
				Type:    session.FlashSuccess,
				Message: "保存しました",
			},
			wantContains: []string{
				`<div id="toaster" class="toaster">`,
				`data-category="success"`,
				`<h2 class="whitespace-pre-line">保存しました</h2>`,
				`<svg`, // successアイコンのSVG
			},
		},
		{
			name: "エラーメッセージ",
			flash: &session.Flash{
				Type:    session.FlashError,
				Message: "エラーが発生しました",
			},
			wantContains: []string{
				`data-category="error"`,
				`<h2 class="whitespace-pre-line">エラーが発生しました</h2>`,
				`<svg`, // errorアイコンのSVG
			},
		},
		{
			name: "警告メッセージ",
			flash: &session.Flash{
				Type:    session.FlashWarning,
				Message: "注意してください",
			},
			wantContains: []string{
				`data-category="warning"`,
				`<h2 class="whitespace-pre-line">注意してください</h2>`,
				`<svg`, // warningアイコンのSVG
			},
		},
		{
			name: "情報メッセージ",
			flash: &session.Flash{
				Type:    session.FlashInfo,
				Message: "お知らせです",
			},
			wantContains: []string{
				`data-category="info"`,
				`<h2 class="whitespace-pre-line">お知らせです</h2>`,
				`<svg`, // infoアイコンのSVG
			},
		},
		{
			name:          "flashがnilの場合は何も表示しない",
			flash:         nil,
			wantNotRender: true,
		},
	}

	for _, tt := range tests {
		tt := tt // キャプチャ
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			ctx = i18n.SetLocale(ctx, "ja") // 日本語ロケールを設定

			// テンプレートをレンダリング
			var buf strings.Builder
			err := Flash(ctx, tt.flash).Render(ctx, &buf)
			if err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			// flashがnilの場合は何も表示されないことを確認
			if tt.wantNotRender {
				if strings.TrimSpace(html) != "" {
					t.Errorf("flashがnilの場合は何も表示されないはずだが、HTMLが生成されました: %s", html)
				}
				return
			}

			// 期待する文字列が含まれているか確認
			for _, want := range tt.wantContains {
				if !strings.Contains(html, want) {
					t.Errorf("期待する文字列が含まれていません: %q\nHTML: %s", want, html)
				}
			}
		})
	}
}

func TestFlash_DismissButton(t *testing.T) {
	t.Parallel()

	flash := &session.Flash{
		Type:    session.FlashSuccess,
		Message: "テスト",
	}

	ctx := context.Background()
	ctx = i18n.SetLocale(ctx, "ja")

	var buf strings.Builder
	err := Flash(ctx, flash).Render(ctx, &buf)
	if err != nil {
		t.Fatalf("レンダリングエラー: %v", err)
	}

	html := buf.String()

	// 閉じるボタンが含まれているか確認
	wantButton := `<button type="button" class="btn" data-toast-action>`
	if !strings.Contains(html, wantButton) {
		t.Errorf("閉じるボタンが含まれていません\nHTML: %s", html)
	}

	// flash_dismissの翻訳が含まれているか確認
	dismissText := i18n.T(ctx, "flash_dismiss")
	if !strings.Contains(html, dismissText) {
		t.Errorf("flash_dismissの翻訳が含まれていません（期待: %q）\nHTML: %s", dismissText, html)
	}
}
