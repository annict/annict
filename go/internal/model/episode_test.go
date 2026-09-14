package model_test

import (
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// TestEpisode_DerivedStatusはtimestampsからstatusへの優先順位の正本を検証する:
// deleted_atがunpublished_atより優先され、どちらのtimestampも無いエピソードはpublished。
func TestEpisode_DerivedStatus(t *testing.T) {
	t.Parallel()

	someTime := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		unpublishedAt *time.Time
		deletedAt     *time.Time
		want          model.EpisodeStatus
	}{
		{"both nil -> published", nil, nil, model.EpisodeStatusPublished},
		{"unpublished_at set -> archived", &someTime, nil, model.EpisodeStatusArchived},
		{"deleted_at set -> deleted", nil, &someTime, model.EpisodeStatusDeleted},
		// deleted_atがunpublished_atより優先される。
		{"both set -> deleted", &someTime, &someTime, model.EpisodeStatusDeleted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &model.Episode{UnpublishedAt: tt.unpublishedAt, DeletedAt: tt.deletedAt}
			if got := e.DerivedStatus(); got != tt.want {
				t.Errorf("DerivedStatus() = %q、期待値 = %q", got, tt.want)
			}
		})
	}
}

// TestManualEpisodeCreationStateは、作品の状態がどの理由を報告するか、および通常の
// コミッターがそのエピソードを作成できるかを検証する。一括作成はAllowedがfalseを返す
// 状態で送信を拒否する。拒否のメッセージもページの警告もRestrictionが返す理由を名指しする
// ため、2つの条件は1つの決まった順序で解決される必要がある。
func TestManualEpisodeCreationState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		state           model.ManualEpisodeCreationState
		wantRestriction model.ManualEpisodeCreationRestriction
		wantAllowed     bool
	}{
		{
			name:            "どちらの条件にも当てはまらない",
			state:           model.ManualEpisodeCreationState{},
			wantRestriction: model.ManualEpisodeCreationAllowed,
			wantAllowed:     true,
		},
		{
			name:            "予定話数まで登録済み",
			state:           model.ManualEpisodeCreationState{EpisodesFilled: true},
			wantRestriction: model.ManualEpisodeCreationEpisodesFilled,
			wantAllowed:     false,
		},
		{
			name:            "放送枠がある",
			state:           model.ManualEpisodeCreationState{SlotsExist: true},
			wantRestriction: model.ManualEpisodeCreationSlotsExist,
			wantAllowed:     false,
		},
		{
			name:            "両方に当てはまるときは予定話数到達を報告する",
			state:           model.ManualEpisodeCreationState{EpisodesFilled: true, SlotsExist: true},
			wantRestriction: model.ManualEpisodeCreationEpisodesFilled,
			wantAllowed:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.state.Restriction(); got != tt.wantRestriction {
				t.Errorf("Restriction() = %q、期待値 = %q", got, tt.wantRestriction)
			}
			if got := tt.state.Allowed(); got != tt.wantAllowed {
				t.Errorf("Allowed() = %v、期待値 = %v", got, tt.wantAllowed)
			}
		})
	}
}
