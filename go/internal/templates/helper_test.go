package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/internal/i18n"
)

// ========================================
// templ用ヘルパー関数のテスト
// ========================================

// TestT は翻訳関数Tが正しく動作することを確認
func TestT(t *testing.T) {
	tests := []struct {
		name      string
		locale    string
		messageID string
		data      []map[string]any
		want      string
	}{
		{
			name:      "日本語の翻訳",
			locale:    i18n.LangJa,
			messageID: "sign_in_heading",
			want:      "ログイン",
		},
		{
			name:      "英語の翻訳",
			locale:    i18n.LangEn,
			messageID: "sign_in_heading",
			want:      "Sign in to Annict",
		},
		{
			name:      "テンプレートデータ付き翻訳",
			locale:    i18n.LangJa,
			messageID: "watchers_count",
			data:      []map[string]any{{"Count": 100}},
			want:      "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := i18n.SetLocale(context.Background(), tt.locale)
			localizer := i18n.GetLocalizer(ctx)
			ctx = i18n.SetLocalizer(ctx, localizer)

			result := T(ctx, tt.messageID, tt.data...)

			if !strings.Contains(result, tt.want) {
				t.Errorf("T() = %q, want to contain %q", result, tt.want)
			}
		})
	}
}

// TestLocale はLocale関数が正しくロケールを返すことを確認
func TestLocale(t *testing.T) {
	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{
			name:   "日本語ロケール",
			locale: i18n.LangJa,
			want:   i18n.LangJa,
		},
		{
			name:   "英語ロケール",
			locale: i18n.LangEn,
			want:   i18n.LangEn,
		},
		{
			name:   "デフォルトロケール",
			locale: "",
			want:   i18n.DefaultLang,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			if tt.locale != "" {
				ctx = i18n.SetLocale(context.Background(), tt.locale)
			} else {
				ctx = context.Background()
			}

			result := Locale(ctx)

			if result != tt.want {
				t.Errorf("Locale() = %q, want %q", result, tt.want)
			}
		})
	}
}

// TestDeref はDeref関数がポインタを正しく参照外しすることを確認
func TestDeref(t *testing.T) {
	t.Run("int32ポインタ", func(t *testing.T) {
		value := int32(2024)
		result := Deref(&value)
		if result != 2024 {
			t.Errorf("Deref() = %d, want 2024", result)
		}
	})

	t.Run("nilポインタ（int32）", func(t *testing.T) {
		var ptr *int32
		result := Deref(ptr)
		if result != 0 {
			t.Errorf("Deref() = %d, want 0", result)
		}
	})

	t.Run("stringポインタ", func(t *testing.T) {
		value := "test"
		result := Deref(&value)
		if result != "test" {
			t.Errorf("Deref() = %q, want \"test\"", result)
		}
	})

	t.Run("nilポインタ（string）", func(t *testing.T) {
		var ptr *string
		result := Deref(ptr)
		if result != "" {
			t.Errorf("Deref() = %q, want empty string", result)
		}
	})

	t.Run("boolポインタ", func(t *testing.T) {
		value := true
		result := Deref(&value)
		if result != true {
			t.Errorf("Deref() = %t, want true", result)
		}
	})

	t.Run("nilポインタ（bool）", func(t *testing.T) {
		var ptr *bool
		result := Deref(ptr)
		if result != false {
			t.Errorf("Deref() = %t, want false", result)
		}
	})
}

// TestIcon はIcon関数が正しいSVGを返すことを確認（templ.Component版）
func TestIcon(t *testing.T) {
	tests := []struct {
		name    string
		icon    string
		class   []string
		wantSVG string
	}{
		{
			name:    "successアイコン",
			icon:    "success",
			wantSVG: `M173.66,98.34`,
		},
		{
			name:    "warningアイコン",
			icon:    "warning",
			wantSVG: `M128,24A104,104`,
		},
		{
			name:    "errorアイコン",
			icon:    "error",
			wantSVG: `M236.8,188.09`,
		},
		{
			name:    "infoアイコン",
			icon:    "info",
			wantSVG: `M128,24A104,104`,
		},
		{
			name:    "sign-inアイコン",
			icon:    "sign-in",
			wantSVG: `M141.66,133.66`,
		},
		{
			name:    "未知のアイコン（infoにフォールバック）",
			icon:    "unknown",
			wantSVG: `M128,24A104,104`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			buf := &bytes.Buffer{}

			// templ.Componentをレンダリング
			component := Icon(tt.icon, tt.class...)
			err := component.Render(ctx, buf)
			if err != nil {
				t.Fatalf("Icon().Render() error = %v", err)
			}

			result := buf.String()

			if !strings.Contains(result, tt.wantSVG) {
				t.Errorf("Icon() does not contain expected SVG fragment %q", tt.wantSVG)
			}

			// SVGタグが含まれることを確認
			if !strings.Contains(result, "<svg") {
				t.Errorf("Icon() does not contain <svg tag")
			}

			// fill="currentColor"を確認
			if !strings.Contains(result, `fill="currentColor"`) {
				t.Errorf("Icon() does not contain fill=\"currentColor\"")
			}
		})
	}
}

// TestIconWithClass はIcon関数のクラス指定機能をテスト
func TestIconWithClass(t *testing.T) {
	tests := []struct {
		name      string
		icon      string
		class     []string
		wantClass string
	}{
		{
			name:      "クラス名あり",
			icon:      "success",
			class:     []string{"fill-green-500"},
			wantClass: `class="fill-green-500"`,
		},
		{
			name:      "複数クラス名",
			icon:      "warning",
			class:     []string{"fill-yellow-500 w-8 h-8"},
			wantClass: `class="fill-yellow-500 w-8 h-8"`,
		},
		{
			name:      "クラス名なし",
			icon:      "info",
			class:     []string{},
			wantClass: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			buf := &bytes.Buffer{}

			component := Icon(tt.icon, tt.class...)
			err := component.Render(ctx, buf)
			if err != nil {
				t.Fatalf("Icon().Render() error = %v", err)
			}

			result := buf.String()

			if tt.wantClass == "" {
				// クラス属性が存在しないことを確認
				if strings.Contains(result, `class=`) {
					t.Errorf("Icon() should not contain class attribute, got: %s", result)
				}
			} else {
				// 指定したクラス属性が含まれることを確認
				if !strings.Contains(result, tt.wantClass) {
					t.Errorf("Icon() does not contain expected class attribute %q, got: %s", tt.wantClass, result)
				}
			}
		})
	}
}
