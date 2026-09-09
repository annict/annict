package model_test

import (
	"testing"
	"time"

	"github.com/annict/annict/go/internal/model"
)

// TestWork_DerivedStatus verifies the single source of truth for the
// timestamp-to-status priority: deleted_at wins over unpublished_at, and a work with
// neither timestamp is published.
//
// [Ja] TestWork_DerivedStatus は timestamps から status への優先順位の正本を検証する:
// deleted_at が unpublished_at より優先され、どちらの timestamp も無い作品は published。
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
		// deleted_at wins over unpublished_at.
		//
		// [Ja] deleted_at が unpublished_at より優先される。
		{"both set -> deleted", &someTime, &someTime, model.WorkStatusDeleted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := &model.Work{UnpublishedAt: tt.unpublishedAt, DeletedAt: tt.deletedAt}
			if got := w.DerivedStatus(); got != tt.want {
				t.Errorf("DerivedStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
