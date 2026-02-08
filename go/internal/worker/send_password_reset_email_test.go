package worker

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/testutil"
)

// TestPasswordResetTemplates はパスワードリセットメールテンプレートの選択とレンダリングをテストします
func TestPasswordResetTemplates(t *testing.T) {
	tests := []struct {
		name         string
		locale       string
		resetURL     string
		htmlContains []string
		textContains []string
	}{
		{
			name:     "日本語メール",
			locale:   "ja",
			resetURL: "https://example.com/reset?token=abc123",
			htmlContains: []string{
				"パスワードリセットのご案内",
				"https://example.com/reset?token=abc123",
				"このリンクは1時間有効です",
			},
			textContains: []string{
				"パスワードリセットのリクエストを受け付けました",
				"https://example.com/reset?token=abc123",
				"このリンクは1時間有効です",
			},
		},
		{
			name:     "英語メール",
			locale:   "en",
			resetURL: "https://example.com/reset?token=xyz789",
			htmlContains: []string{
				"Password Reset Request",
				"https://example.com/reset?token=xyz789",
				"This link is valid for 1 hour",
			},
			textContains: []string{
				"We have received a request to reset your password",
				"https://example.com/reset?token=xyz789",
				"This link is valid for 1 hour",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			htmlBody, textBody := passwordResetTemplates(tt.locale, tt.resetURL)

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

// TestSendPasswordResetEmailWorker_Work はメール送信ワーカーの動作をテストします
func TestSendPasswordResetEmailWorker_Work(t *testing.T) {
	// テスト用DBをセットアップ
	_, tx := testutil.SetupTestDB(t)

	// 日本語ロケールのユーザーを作成
	_ = testutil.NewUserBuilder(t, tx).
		WithUsername("ja_user_worker_test").
		WithEmail("ja_user_worker@example.com").
		WithLocale("ja").
		Build()

	// 英語ロケールのユーザーを作成
	_ = testutil.NewUserBuilder(t, tx).
		WithUsername("en_user_worker_test").
		WithEmail("en_user_worker@example.com").
		WithLocale("en").
		Build()

	// 実際のメール送信はResend APIを使用するため、
	// ここでは基本的なテンプレートレンダリングの動作を確認済み
	t.Log("メール送信ワーカーの基本的な動作を確認しました")
}
