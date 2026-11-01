package model

import "testing"

// TestEpisodeGenerationReport_IsEmptyは、作ったエピソード・飛ばした放送枠・エラーのいずれかが
// あるときだけ報告することがあるとし、件数だけの項目は見ないことを検証する。
func TestEpisodeGenerationReport_IsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		report EpisodeGenerationReport
		want   bool
	}{
		{name: "何も無い", report: EpisodeGenerationReport{}, want: true},
		{name: "紐づけた放送枠とタイトルを入れたエピソードの件数だけ", report: EpisodeGenerationReport{LinkedSlotCount: 2, FilledTitleCount: 3}, want: true},
		{name: "作ったエピソード", report: EpisodeGenerationReport{CreatedEpisodes: []EpisodeGenerationCreatedEpisode{{WorkID: 1}}}, want: false},
		{name: "飛ばした放送枠", report: EpisodeGenerationReport{SkippedSlots: []EpisodeGenerationSkippedSlot{{SlotID: 1}}}, want: false},
		{name: "エラー", report: EpisodeGenerationReport{Errors: []EpisodeGenerationError{{Message: "失敗"}}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.report.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v、期待値 = %v", got, tt.want)
			}
		})
	}
}
