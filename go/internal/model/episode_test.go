package model_test

import (
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// TestEpisode_DerivedStatus verifies the single source of truth for the
// timestamp-to-status priority: deleted_at wins over unpublished_at, and an episode
// with neither timestamp is published.
//
// [Ja] TestEpisode_DerivedStatus は timestamps から status への優先順位の正本を検証する:
// deleted_at が unpublished_at より優先され、どちらの timestamp も無いエピソードは published。
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
		// deleted_at wins over unpublished_at.
		//
		// [Ja] deleted_at が unpublished_at より優先される。
		{"both set -> deleted", &someTime, &someTime, model.EpisodeStatusDeleted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &model.Episode{UnpublishedAt: tt.unpublishedAt, DeletedAt: tt.deletedAt}
			if got := e.DerivedStatus(); got != tt.want {
				t.Errorf("DerivedStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestManualEpisodeCreationState verifies which reason a work's state reports and whether
// ordinary committers may create its episodes. The bulk create refuses a submit when Allowed
// reports false, and both the refusal message and the page's warning name the reason Restriction
// returns, so the two conditions have to resolve in one fixed order.
//
// [Ja] TestManualEpisodeCreationState は、作品の状態がどの理由を報告するか、および通常の
// コミッターがそのエピソードを作成できるかを検証する。一括作成は Allowed が false を返す
// 状態で送信を拒否する。拒否のメッセージもページの警告も Restriction が返す理由を名指しする
// ため、2 つの条件は 1 つの決まった順序で解決される必要がある。
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
				t.Errorf("Restriction() = %q, want %q", got, tt.wantRestriction)
			}
			if got := tt.state.Allowed(); got != tt.wantAllowed {
				t.Errorf("Allowed() = %v, want %v", got, tt.wantAllowed)
			}
		})
	}
}
