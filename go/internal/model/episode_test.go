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
