package model_test

import (
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// TestNumberFormat_EpisodeNumberは、話数を書式に当てはめた表示の番号がRailsの
// Work#formatted_numberと同じになることを検証する。
func TestNumberFormat_EpisodeNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		format    model.NumberFormat
		rawNumber float64
		want      string
	}{
		{"書式に話数を埋め込む", model.NumberFormat{Format: "第%d話"}, 3, "第3話"},
		{"ゼロ埋めの書式", model.NumberFormat{Format: "#%02d"}, 7, "#07"},
		{"小数の話数は整数部だけを使う", model.NumberFormat{Format: "第%d話"}, 3.5, "第3話"},
		{"書式が空なら話数番目の表示を引く", model.NumberFormat{Data: []string{"第一話", "第二話", "第三話"}}, 2, "第二話"},
		{"表示の数を超える話数は空文字列", model.NumberFormat{Data: []string{"第一話", "第二話"}}, 3, ""},
		{"0話以下は空文字列", model.NumberFormat{Data: []string{"第一話", "第二話"}}, 0, ""},
		{"指定子の無い書式はそのまま返す", model.NumberFormat{Format: "特別編"}, 1, "特別編"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.format.EpisodeNumber(tt.rawNumber); got != tt.want {
				t.Errorf("EpisodeNumber(%v) = %q、期待値 = %q", tt.rawNumber, got, tt.want)
			}
		})
	}
}
