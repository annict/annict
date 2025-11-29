package sign_in_code

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/internal/i18n"
)

func TestSignInCodeShow_BackURLHiddenField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		backURL string
		want    string
	}{
		{
			name:    "空のbackURL",
			backURL: "",
			want:    `<input type="hidden" name="back" value="">`,
		},
		{
			name:    "単純なパス",
			backURL: "/works",
			want:    `<input type="hidden" name="back" value="/works">`,
		},
		{
			name:    "クエリパラメータ付きパス",
			backURL: "/oauth/authorize?client_id=xxx",
			want:    `<input type="hidden" name="back" value="/oauth/authorize?client_id=xxx">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// コンテキストにロケールを設定
			ctx := i18n.SetLocale(context.Background(), "ja")

			// テンプレートをレンダリング
			var buf bytes.Buffer
			component := SignInCodeShow(ctx, "test@example.com", nil, "test-csrf-token", tt.backURL)
			err := component.Render(ctx, &buf)
			if err != nil {
				t.Fatalf("テンプレートレンダリングエラー: %v", err)
			}

			html := buf.String()

			// backのhiddenフィールドが含まれていることを確認
			if !strings.Contains(html, tt.want) {
				t.Errorf("backのhiddenフィールドが見つかりません\nexpected to contain: %s\ngot: %s", tt.want, html)
			}
		})
	}
}

func TestSignInCodeShow_BackURLInResendForm(t *testing.T) {
	t.Parallel()

	// コンテキストにロケールを設定
	ctx := i18n.SetLocale(context.Background(), "ja")
	backURL := "/oauth/authorize?client_id=test"

	// テンプレートをレンダリング
	var buf bytes.Buffer
	component := SignInCodeShow(ctx, "test@example.com", nil, "test-csrf-token", backURL)
	err := component.Render(ctx, &buf)
	if err != nil {
		t.Fatalf("テンプレートレンダリングエラー: %v", err)
	}

	html := buf.String()

	// メインフォームと再送信フォームの両方にbackフィールドが含まれていることを確認
	// 2つのフォームがあるので、backフィールドは2回出現するはず
	backFieldCount := strings.Count(html, `name="back"`)
	if backFieldCount != 2 {
		t.Errorf("backフィールドの数が期待と異なります: got %d, want 2", backFieldCount)
	}
}
