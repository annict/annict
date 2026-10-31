package model_test

import (
	"testing"

	"github.com/annict/annict/go/internal/model"
)

// TestSlot_EpisodeRawNumberは、放送枠が求める話数を、回数から前にある不定期放送の数を
// 引き、作品の開始話数を起点にして数えることを検証する。
func TestSlot_EpisodeRawNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		slot                  model.Slot
		startEpisodeRawNumber float64
		wantOrder             int32
		wantRawNumber         float64
	}{
		{"不定期放送の無い番組", model.Slot{Number: 3}, 1, 3, 3},
		{"前の不定期放送を引く", model.Slot{Number: 5, PrecedingIrregularCount: 2}, 1, 3, 3},
		{"開始話数が13の作品", model.Slot{Number: 2}, 13, 2, 14},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.slot.EpisodeOrder(); got != tt.wantOrder {
				t.Errorf("EpisodeOrder() = %d、期待値 = %d", got, tt.wantOrder)
			}
			if got := tt.slot.EpisodeRawNumber(tt.startEpisodeRawNumber); got != tt.wantRawNumber {
				t.Errorf("EpisodeRawNumber(%v) = %v、期待値 = %v", tt.startEpisodeRawNumber, got, tt.wantRawNumber)
			}
		})
	}
}
