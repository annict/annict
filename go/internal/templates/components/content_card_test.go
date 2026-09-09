package components

import (
	"context"
	"strings"
	"testing"
)

// TestContentCard_Class verifies the optional class argument is appended to the card's
// class list, and that an empty argument leaves the base classes untouched.
//
// [Ja] TestContentCard_Class は任意の class 引数がカードのクラスリストへ追記されること、
// 空引数のときは基底クラスがそのまま残ることを検証する。
func TestContentCard_Class(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		class    string
		expected string
	}{
		{
			name:     "空引数では基底クラスのみ",
			class:    "",
			expected: `class="card py-4 rounded-none md:rounded-xl mx-0 md:mx-4"`,
		},
		{
			// overflow-visible lets an inner combobox popover escape the card's default
			// overflow-hidden clipping.
			//
			// [Ja] overflow-visible により、内側の combobox popover がカード既定の
			// overflow-hidden によるクリップから逃れられる。
			name:     "overflow-visible が追記される",
			class:    "overflow-visible",
			expected: `class="card py-4 rounded-none md:rounded-xl mx-0 md:mx-4 overflow-visible"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			if err := ContentCard(tt.class).Render(context.Background(), &buf); err != nil {
				t.Fatalf("レンダリングエラー: %v", err)
			}

			if html := buf.String(); !strings.Contains(html, tt.expected) {
				t.Errorf("期待する class が含まれていません: %q\n出力: %s", tt.expected, html)
			}
		})
	}
}
