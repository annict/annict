package model_test

import (
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// TestWork_DerivedStatusはtimestampsからstatusへの優先順位の正本を検証する:
// deleted_atがunpublished_atより優先され、どちらのtimestampも無い作品はpublished。
func TestWork_DerivedStatus(t *testing.T) {
	t.Parallel()

	someTime := time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		unpublishedAt *time.Time
		deletedAt     *time.Time
		want          model.WorkStatus
	}{
		{"both nil -> published", nil, nil, model.WorkStatusPublished},
		{"unpublished_at set -> archived", &someTime, nil, model.WorkStatusArchived},
		{"deleted_at set -> deleted", nil, &someTime, model.WorkStatusDeleted},
		// deleted_atがunpublished_atより優先される。
		{"both set -> deleted", &someTime, &someTime, model.WorkStatusDeleted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &model.Work{UnpublishedAt: tt.unpublishedAt, DeletedAt: tt.deletedAt}
			if got := w.DerivedStatus(); got != tt.want {
				t.Errorf("DerivedStatus() = %q、期待値 = %q", got, tt.want)
			}
		})
	}
}
