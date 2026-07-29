package components

import (
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
)

// TestRequiredBadge verifies that the marker names the requirement in words in each locale,
// so it carries meaning without a legend and reads as words to a screen reader.
//
// [Ja] TestRequiredBadge は、印が各ロケールで必須であることを言葉で示すことを検証する。
// これにより凡例が無くても意味を持ち、スクリーンリーダーにも言葉として読まれる。
func TestRequiredBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{
			name:   "日本語",
			locale: "ja",
			want:   "必須",
		},
		{
			name:   "英語",
			locale: "en",
			want:   "Required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := i18n.SetLocale(context.Background(), tt.locale)

			var buf strings.Builder
			if err := RequiredBadge().Render(ctx, &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			html := buf.String()

			if !strings.Contains(html, tt.want) {
				t.Errorf("期待する文字列が含まれていません: %q, got %q", tt.want, html)
			}

			// The outline badge is what gives the marker its foreground color. Pinning the
			// variant keeps the marker from drifting to one whose color reads as a status.
			//
			// [Ja] 印の文字色は outline の Badge が与えている。variant を固定することで、
			// 色が状態を示すように読める別の variant へ流れるのを防ぐ。
			if !strings.Contains(html, `class="badge" data-variant="outline"`) {
				t.Errorf("必須の印は outline variant の Badge で表示するべきです: %q", html)
			}

			// The destructive color belongs to validation errors; being required is a normal
			// state and must not borrow the same signal.
			//
			// [Ja] destructive の色はバリデーションエラーのものであり、平常の状態である必須が
			// 同じ合図を借りてはいけない。
			if strings.Contains(html, "text-destructive") {
				t.Error("必須の印はエラー用の色を使ってはいけません")
			}
		})
	}
}
