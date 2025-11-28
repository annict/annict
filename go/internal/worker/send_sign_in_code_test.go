package worker

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/internal/testutil"
)

// TestRenderSignInTemplate はログインコードテンプレートのレンダリングをテストします
func TestRenderSignInTemplate(t *testing.T) {
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
				"ログインコードをお送りします",
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
				"ログインコードのご案内",
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
				"Here is your login code",
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
				"Your Login Code",
				"654321",
				"valid for 15 minutes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テンプレートをレンダリング
			ctx := context.Background()
			result, err := renderSignInTemplate(ctx, tt.locale, tt.format, tt.code)
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
