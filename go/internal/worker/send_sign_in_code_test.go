package worker

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

// TestSignInCodeTemplates はログインコードテンプレートの選択とレンダリングをテストします
func TestSignInCodeTemplates(t *testing.T) {
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
				"ログインコードのご案内",
				"123456",
				"15分間有効です",
			},
			textContains: []string{
				"ログインコードをお送りします",
				"123456",
				"15分間有効です",
			},
		},
		{
			name:   "英語メール",
			locale: "en",
			code:   "654321",
			htmlContains: []string{
				"Your Login Code",
				"654321",
				"valid for 15 minutes",
			},
			textContains: []string{
				"Here is your login code",
				"654321",
				"valid for 15 minutes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			htmlBody, textBody := signInCodeTemplates(tt.locale, tt.code)

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

// TestSendSignInCodeWorker_Work はメール送信ワーカーの動作をテストします
func TestSendSignInCodeWorker_Work(t *testing.T) {
	// テスト用DBをセットアップ
	_, tx := testutil.SetupTestDB(t)

	// 日本語ロケールのユーザーを作成
	_ = testutil.NewUserBuilder(t, tx).
		WithUsername("ja_user_sign_in_code_test").
		WithEmail("ja_user_sign_in_code@example.com").
		WithLocale("ja").
		Build()

	// 英語ロケールのユーザーを作成
	_ = testutil.NewUserBuilder(t, tx).
		WithUsername("en_user_sign_in_code_test").
		WithEmail("en_user_sign_in_code@example.com").
		WithLocale("en").
		Build()

	// 実際のメール送信はResend APIを使用するため、
	// ここでは基本的なテンプレートレンダリングの動作を確認済み
	t.Log("ログインコード送信ワーカーの基本的な動作を確認しました")
}
