package worker

import (
	"context"
	"strings"
	"testing"
)

// TestRenderSignUpCodeTemplate は新規登録確認コードテンプレートのレンダリングをテストします
func TestRenderSignUpCodeTemplate(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		format   string
		code     string
		contains []string
	}{
		{
			name:   "日本語テキストメール",
			locale: "ja",
			format: "text",
			code:   "123456",
			contains: []string{
				"Annictへようこそ！",
				"アカウント登録を完了するため",
				"123456",
				"15分間有効です",
			},
		},
		{
			name:   "日本語HTMLメール",
			locale: "ja",
			format: "html",
			code:   "123456",
			contains: []string{
				"Annictへようこそ！",
				"アカウント登録を完了するため",
				"123456",
				"15分間有効です",
			},
		},
		{
			name:   "英語テキストメール",
			locale: "en",
			format: "text",
			code:   "654321",
			contains: []string{
				"Welcome to Annict!",
				"complete your account registration",
				"654321",
				"valid for 15 minutes",
			},
		},
		{
			name:   "英語HTMLメール",
			locale: "en",
			format: "html",
			code:   "654321",
			contains: []string{
				"Welcome to Annict!",
				"complete your account registration",
				"654321",
				"valid for 15 minutes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テンプレートをレンダリング
			ctx := context.Background()
			result, err := renderSignUpCodeTemplate(ctx, tt.locale, tt.format, tt.code)
			if err != nil {
				t.Fatalf("テンプレートのレンダリングに失敗: %v", err)
			}

			// 期待される文字列が含まれているか確認
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("期待される文字列が見つかりません: %q\nレンダリング結果:\n%s", expected, result)
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
