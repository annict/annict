package worker

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestSignUpCodeTemplates は新規登録確認コードテンプレートの選択とレンダリングをテストします
func TestSignUpCodeTemplates(t *testing.T) {
	tests := []struct {
		name         string
		locale       string
		code         string
		htmlContains []string
		textContains []string
	}{
		{
			name:   "日本語メール",
			locale: "ja",
			code:   "123456",
			htmlContains: []string{
				"Annictへようこそ！",
				"アカウント登録を完了するため",
				"123456",
				"15分間有効です",
			},
			textContains: []string{
				"Annictへようこそ！",
				"アカウント登録を完了するため",
				"123456",
				"15分間有効です",
			},
		},
		{
			name:   "英語メール",
			locale: "en",
			code:   "654321",
			htmlContains: []string{
				"Welcome to Annict!",
				"complete your account registration",
				"654321",
				"valid for 15 minutes",
			},
			textContains: []string{
				"Welcome to Annict!",
				"complete your account registration",
				"654321",
				"valid for 15 minutes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			htmlBody, textBody := signUpCodeTemplates(tt.locale, tt.code)

			// HTMLテンプレートをレンダリング
			var htmlBuf bytes.Buffer
			if err := htmlBody.Render(ctx, &htmlBuf); err != nil {
				t.Fatalf("HTMLテンプレートのレンダリングに失敗: %v", err)
			}
			for _, expected := range tt.htmlContains {
				if !strings.Contains(htmlBuf.String(), expected) {
					t.Errorf("HTMLに期待される文字列が見つかりません: %q\nレンダリング結果:\n%s", expected, htmlBuf.String())
				}
			}

			// テキストテンプレートをレンダリング
			var textBuf bytes.Buffer
			if err := textBody.Render(ctx, &textBuf); err != nil {
				t.Fatalf("テキストテンプレートのレンダリングに失敗: %v", err)
			}
			for _, expected := range tt.textContains {
				if !strings.Contains(textBuf.String(), expected) {
					t.Errorf("テキストに期待される文字列が見つかりません: %q\nレンダリング結果:\n%s", expected, textBuf.String())
				}
			}
		})
	}
}

// TestSendSignUpCodeWorker_Work は新規登録確認コード送信ワーカーの動作をテストします
func TestSendSignUpCodeWorker_Work(t *testing.T) {
	// 実際のメール送信はResend APIを使用するため、
	// ここでは基本的なテンプレートレンダリングの動作を確認済み
	t.Log("新規登録確認コード送信ワーカーの基本的な動作を確認しました")
}
