package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/annict/annict/go/internal/i18n"
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
			icon:    "sign-in-regular",
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

// TestLabeledIcon verifies that a meaningful SVG has image semantics, a safely escaped
// accessible label, and no place in the legacy SVG focus order.
//
// [Ja] TestLabeledIcon は意味を持つ SVG に画像のセマンティクスと安全にエスケープされた
// アクセシブルネームが付き、従来の SVG フォーカス順序から除外されることを検証する。
func TestLabeledIcon(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := LabeledIcon("info", `Episode "1" & details`, "size-4").Render(context.Background(), &buf); err != nil {
		t.Fatalf("LabeledIcon().Render() error = %v", err)
	}

	result := buf.String()
	expectedFragments := []string{
		`<svg role="img" aria-label="Episode &#34;1&#34; &amp; details" focusable="false" class="size-4"`,
		`M128,24A104,104`,
	}
	for _, expected := range expectedFragments {
		if !strings.Contains(result, expected) {
			t.Errorf("LabeledIcon() does not contain expected fragment %q, got: %s", expected, result)
		}
	}

	if strings.Contains(result, `aria-hidden="true"`) {
		t.Errorf("LabeledIcon() should remain exposed to assistive technology, got: %s", result)
	}
}

// TestDecorativeIcon verifies that a decorative SVG is hidden from assistive technology and
// excluded from the legacy SVG focus order while retaining the requested presentation class.
//
// [Ja] TestDecorativeIcon は装飾 SVG が支援技術から隠れ、従来の SVG フォーカス順序から
// 除外されつつ、指定した表示用 class を保つことを検証する。
func TestDecorativeIcon(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := DecorativeIcon("plus-regular", "size-4").Render(context.Background(), &buf); err != nil {
		t.Fatalf("DecorativeIcon().Render() error = %v", err)
	}

	result := buf.String()
	expectedFragments := []string{
		`<svg aria-hidden="true" focusable="false" class="size-4"`,
		`M224,128`,
	}
	for _, expected := range expectedFragments {
		if !strings.Contains(result, expected) {
			t.Errorf("DecorativeIcon() does not contain expected fragment %q, got: %s", expected, result)
		}
	}

	if strings.Contains(result, "<span") {
		t.Errorf("DecorativeIcon() should apply attributes directly to the SVG, got: %s", result)
	}
}

// TestDecorativeInlineIcon verifies the two positions Basecoat documents and rejects any
// other value instead of copying an unchecked data-icon attribute into the SVG. Every case
// remains decorative and outside the legacy SVG focus order.
//
// [Ja] TestDecorativeInlineIcon は Basecoat が文書化する 2 つの位置を検証し、それ以外の値を
// 未検証の data-icon 属性として SVG にコピーせず拒否することを確認する。どのケースも装飾の
// ままで、従来の SVG フォーカス順序から除外される。
func TestDecorativeInlineIcon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		position      InlineIconPosition
		wantAttribute string
	}{
		{name: "先頭", position: InlineIconStart, wantAttribute: `data-icon="inline-start"`},
		{name: "末尾", position: InlineIconEnd, wantAttribute: `data-icon="inline-end"`},
		{name: "不明な位置", position: InlineIconPosition("outside"), wantAttribute: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			if err := DecorativeInlineIcon("plus-regular", tt.position, "size-4").Render(context.Background(), &buf); err != nil {
				t.Fatalf("DecorativeInlineIcon().Render() error = %v", err)
			}

			result := buf.String()
			for _, expected := range []string{`aria-hidden="true"`, `focusable="false"`, `class="size-4"`, `M224,128`} {
				if !strings.Contains(result, expected) {
					t.Errorf("DecorativeInlineIcon() does not contain expected fragment %q, got: %s", expected, result)
				}
			}

			if tt.wantAttribute == "" {
				if strings.Contains(result, "data-icon=") {
					t.Errorf("DecorativeInlineIcon() should omit an unknown position, got: %s", result)
				}
			} else if !strings.Contains(result, tt.wantAttribute) {
				t.Errorf("DecorativeInlineIcon() does not contain expected position %q, got: %s", tt.wantAttribute, result)
			}
		})
	}
}

// TestPhosphorIconsStartWithSVGTag holds the invariant iconSVG relies on when it adds
// attributes: every stored icon begins with the literal "<svg " that the helpers replace.
// An icon added in any other shape would render broken markup rather than fail, so the
// assertion covers the whole map instead of the icons that happen to be used today.
//
// [Ja] TestPhosphorIconsStartWithSVGTag は、iconSVG が属性を足すときに前提としている不変条件
// (保持している各アイコンが、ヘルパーの置き換え対象である "<svg " のリテラルで始まること) を
// 担保する。別の形で追加されたアイコンは失敗せず壊れたマークアップを描画するため、現に使われて
// いるアイコンだけでなく map 全体を対象にする。
func TestPhosphorIconsStartWithSVGTag(t *testing.T) {
	t.Parallel()

	for name, svg := range phosphorIcons {
		if !strings.HasPrefix(svg, "<svg ") {
			t.Errorf("phosphorIcons[%q] should start with %q, got: %.20s", name, "<svg ", svg)
		}
	}
}

// TestPhosphorIconsHoldFallbackIcon holds the other invariant iconSVG relies on: the map has
// an entry for the name unknown icons fall back to. Without it the fallback resolves to an
// empty string, and taking the leading "<svg " off it panics instead of rendering a
// placeholder. Icon reached that path only when a class was passed, but DecorativeIcon and
// LabeledIcon take it on every call.
//
// [Ja] TestPhosphorIconsHoldFallbackIcon は iconSVG が前提とするもう 1 つの不変条件
// (未知のアイコンがフォールバックする名前のエントリを map が持つこと) を担保する。無いと
// フォールバックが空文字列に解決し、先頭の "<svg " を取り除く処理が代替を描画せず panic する。
// Icon はこの経路を class を渡したときにしか通らなかったが、DecorativeIcon と LabeledIcon は
// 毎回通る。
func TestPhosphorIconsHoldFallbackIcon(t *testing.T) {
	t.Parallel()

	if _, ok := phosphorIcons[fallbackIconName]; !ok {
		t.Fatalf("phosphorIcons はフォールバック先の %q を持つ必要があります", fallbackIconName)
	}
}
