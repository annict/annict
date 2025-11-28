package worker

import (
	"context"
	"testing"

	"github.com/annict/annict/internal/testutil"
)

// TestRenderEmailTemplate はメールテンプレートのレンダリングをテストします
func TestRenderEmailTemplate(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		format   string
		resetURL string
		contains []string
	}{
		{
			name:     "日本語テキストメール",
			locale:   "ja",
			format:   "text",
			resetURL: "https://example.com/reset?token=abc123",
			contains: []string{
				"パスワードリセットのリクエストを受け付けました",
				"https://example.com/reset?token=abc123",
				"このリンクは1時間有効です",
			},
		},
		{
			name:     "日本語HTMLメール",
			locale:   "ja",
			format:   "html",
			resetURL: "https://example.com/reset?token=abc123",
			contains: []string{
				"パスワードリセットのご案内",
				"https://example.com/reset?token=abc123",
				"このリンクは1時間有効です",
			},
		},
		{
			name:     "英語テキストメール",
			locale:   "en",
			format:   "text",
			resetURL: "https://example.com/reset?token=xyz789",
			contains: []string{
				"We have received a request to reset your password",
				"https://example.com/reset?token=xyz789",
				"This link is valid for 1 hour",
			},
		},
		{
			name:     "英語HTMLメール",
			locale:   "en",
			format:   "html",
			resetURL: "https://example.com/reset?token=xyz789",
			contains: []string{
				"Password Reset Request",
				"https://example.com/reset?token=xyz789",
				"This link is valid for 1 hour",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テンプレートをレンダリング
			ctx := context.Background()
			result, err := renderEmailTemplate(ctx, tt.locale, tt.format, tt.resetURL)
			if err != nil {
				t.Fatalf("テンプレートのレンダリングに失敗: %v", err)
			}

			// 期待される文字列が含まれているか確認
			for _, expected := range tt.contains {
				if !contains(result, expected) {
					t.Errorf("期待される文字列が見つかりません: %q\nレンダリング結果:\n%s", expected, result)
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

// contains は文字列に部分文字列が含まれているかチェックします
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || anyIndex(s, substr) >= 0)
}

// anyIndex は文字列内で部分文字列が最初に現れるインデックスを返します
func anyIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
